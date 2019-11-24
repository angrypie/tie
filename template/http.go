package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

func makeHTTPServer(info *PackageInfo, main *Group, f *File) {
	service := info.Service
	if service.Type != "http" && service.Type != "httpOnly" {
		return
	}
	makeStartHTTPServer(info, main, f)
	makeHTTPRequestResponseTypes(info, main, f)
	makeHTTPHandlers(info, main, f)

	makeHelpers(info, main, f)

}

func makeHTTPHandlers(info *PackageInfo, main *Group, f *File) {
	f.Comment(fmt.Sprintf("API handler methods (%s)", "HTTP")).Line()
	forEachFunction(info.Functions, true, func(fn *parser.Function) {
		makeHTTPHandler(info, fn, f.Group)
	})
}

func makeHTTPHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	echo := "github.com/labstack/echo"
	handler, request, response := getMethodTypes(fn, "HTTP")
	receiverVarName := getReceiverVarName(fn.Receiver.Type)
	handlerBody := func(g *Group) {
		//Bind request params
		if len(fn.Arguments) > 0 {
			g.Id("request").Op(":=").New(Id(request))
			g.List(Id("_"), ListFunc(createArgsListFunc(fn.Arguments, "request", "string,"))).Op("=").
				List(Lit(0), ListFunc(createArgsList(fn.Arguments, func(arg *Statement) *Statement {
					return Id(firstNotEmptyStrHelper).Call(
						Id("request").Dot(arg.GoString()),
						Id("ctx").Dot("QueryParam").Call(Lit(strings.ToLower(arg.GoString()))),
					)
				}, "", "string,")))
			g.If(
				Err().Op(":=").Id("ctx").Dot("Bind").Call(Id("request")),
				Err().Op("!=").Nil(),
			).Block(Return(Err()))
			//Empty argument needs to avoid errors if no other arguments exist
		}

		//Call original function
		g.Id("response").Op(":=").New(Id(response))

		//If method has receiver generate receiver middleware code
		//else just call public package method
		if hasReceiver(fn) {
			initReceiverFunc := findInitReceiver(info.Functions, fn)
			if initReceiverFunc != nil && !isTopLevelInitReceiver(initReceiverFunc) {
				receiverType := fn.Receiver.Type
				g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, strings.Trim(receiverType, "*")).Block()
				makeReceiverMiddleware(receiverVarName, g, initReceiverFunc, info)
			}
			injectOriginalMethodCall(g, fn, Id(receiverVarName).Dot(fn.Name))
		} else {
			injectOriginalMethodCall(g, fn, Qual(info.Service.Name, fn.Name))
		}

		ifErrorReturnBadRequestWithErr(
			g,
			Err().Op(":=").Id("response").Dot("Err"),
		)

		g.Return(Id("ctx").Dot("JSON").Call(Qual("net/http", "StatusOK"), Id("response")))
	}

	//Create handler methods that use closure to inject receiver if it exist.
	if hasReceiver(fn) {
		file.Func().Id(handler).ParamsFunc(func(g *Group) {
			initReceiverFunc := findInitReceiver(info.Functions, fn)
			if initReceiverFunc == nil || isTopLevelInitReceiver(initReceiverFunc) {
				g.Id(receiverVarName).Op("*").Qual(info.Service.Name, strings.Trim(fn.Receiver.Type, "*"))
			} else {
				g.Add(getInitReceiverDepsSignature(initReceiverFunc, info))
			}
		}).Params(
			Func().Params(Qual(echo, "Context")).Params(Error()),
		).Block(Return(Func().
			Params(Id("ctx").Qual(echo, "Context")).
			Params(Err().Error()).BlockFunc(handlerBody),
		)).Line()
	} else {
		file.Func().Id(handler).
			Params(Id("ctx").Qual(echo, "Context")).
			Params(Err().Error()).BlockFunc(handlerBody).
			Line()
	}

}

func makeHTTPRequestResponseTypes(info *PackageInfo, main *Group, f *File) {
	f.Add(createReqRespTypes("HTTP", info))
}

func makeStartHTTPServer(info *PackageInfo, main *Group, f *File) {
	echoMiddleware := "github.com/labstack/echo/middleware"
	rndport := "github.com/angrypie/rndport"
	echo := "github.com/labstack/echo"

	main.Go().Id("startHTTPServer").Call()

	f.Func().Id("startHTTPServer").Params().BlockFunc(func(g *Group) {
		//Declare err and get rid of ''unused' error.
		g.Var().Err().Error()
		g.Id("_").Op("=").Err()

		port := info.Service.Port
		if port == "" {
			g.List(Id("address"), Err()).Op(":=").
				Qual(rndport, "GetAddress").Call(Lit(":%d"))
			g.If(Err().Op("!=").Nil()).Block(Panic(Err()))
		} else {
			g.Id("address").Op(":=").Lit(fmt.Sprintf(":%s", port))
		}

		g.Id("server").Op(":=").Qual(echo, "New").Call()
		g.Id("server").Dot("Use").Call(Qual(echoMiddleware, "CORSWithConfig").Call(
			Qual(echoMiddleware, "CORSConfig").Values(Dict{
				Id("AllowOrigins"): Index().String().Values(Lit("*")),
			}),
		))

		//. Set HTTP handlers and init receivers.
		//.1 Create receivers for handlers
		receiversCreated := make(map[string]*parser.Function)
		forEachFunction(info.Functions, false, func(fn *parser.Function) {
			receiverType := fn.Receiver.Type
			//Next function if receiver already created
			if !hasReceiver(fn) || receiversCreated[receiverType] != nil {
				return
			}
			initReceiverFunc := findInitReceiver(info.Functions, fn)
			if initReceiverFunc != nil && !isTopLevelInitReceiver(initReceiverFunc) {
				return
			}
			receiverVarName := getReceiverVarName(receiverType)
			g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, strings.Trim(receiverType, "*")).Block()
			makeReceiverInitialization(receiverVarName, g, initReceiverFunc, info)
			receiversCreated[receiverType] = initReceiverFunc
		})
		//.2 Add http handler for each function.
		//Route format is /receiver_name/method_name
		forEachFunction(info.Functions, true, func(fn *parser.Function) {
			handler, _, _ := getMethodTypes(fn, "HTTP")

			//If handler has receiver
			if hasReceiver(fn) {
				initReceiverFunc := findInitReceiver(info.Functions, fn)
				receiverType := fn.Receiver.Type
				receiverVarName := getReceiverVarName(receiverType)
				route := fmt.Sprintf("%s/%s", receiverType, fn.Name)

				g.Id("server").Dot("POST").Call(
					Lit(toSnakeCase(route)),
					Id(handler).CallFunc(func(g *Group) {
						if initReceiverFunc == nil || isTopLevelInitReceiver(initReceiverFunc) {
							g.Id(receiverVarName)
						} else {
							g.Add(getInitReceiverDepsNames(initReceiverFunc))
						}
					}),
				)
				return
			}

			route := fn.Name
			g.Id("server").Dot("POST").Call(
				Lit(toSnakeCase(route)),
				Id(handler),
			)
		})

		if key := info.Service.Auth; key != "" {
			g.Id("server").Dot("Use").Call(Qual(echoMiddleware, "KeyAuth").Call(
				Func().Params(Id("key").String(), Id("ctx").Qual(echo, "Context")).Params(Bool(), Error()).Block(
					Id("auth").Op(":=").Lit(key),
					If(
						Id("envKey").Op(":=").Id(getEnvHelper).Call(Lit("TIE_API_KEY")),
						Id("envKey").Op("!=").Lit(""),
					).Block(
						Id("auth").Op("=").Id("envKey"),
					),
					Return().List(Id("key").Op("==").Id("auth"), Nil()),
				)),
			)
		}
		g.Id("server").Dot("Start").Call(Id("address"))

	})
}

func makeReceiverMiddleware(recId string, scope *Group, initReceiver *parser.Function, info *PackageInfo) {
	if initReceiver == nil {
		return
	}
	initReceiverCall := func(g *Group) {
		for _, field := range initReceiver.Arguments {
			name := field.Name
			//TODO check getHeader and getEnv function signature
			//Inject getHeader function that returns header of current request
			if name == "getHeader" {
				headerArg := "header"
				g.Func().Params(Id(headerArg).String()).String().Block(
					Return(Id("ctx").Dot("Request").Call().Dot("Header").Dot("Get").Call(Id(headerArg))),
				)
				continue
			}
			//Inject getEnv function that provide access to environment variables
			if name == "getEnv" {
				g.Id(getEnvHelper)
				continue
			}

			//Oterwise inject receiver dependencie
			injectReceiverName(g, field)
		}
	}

	ifErrorReturnBadRequestWithErr(
		scope,
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, initReceiver.Name).CallFunc(initReceiverCall),
	)
}

func makeReceiverInitialization(recId string, scope *Group, initReceiver *parser.Function, info *PackageInfo) {
	if initReceiver == nil {
		return
	}
	initReceiverCall := func(g *Group) {
		for _, field := range initReceiver.Arguments {
			name := field.Name
			//TODO check getEnv function signature
			//Inject getEnv function that provide access to environment variables
			if name == "getEnv" {
				g.Id(getEnvHelper)
			}
		}
	}

	scope.If(
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, initReceiver.Name).CallFunc(initReceiverCall),
		Err().Op("!=").Nil(),
	).Block(
		//TODO return appropriate error here
		Panic(Err()),
	)

}

func ifErrorReturnBadRequestWithErr(scope *Group, statement *Statement) {
	scope.If(
		statement,
		Err().Op("!=").Nil(),
	).Block(
		Return(Id("ctx").Dot("JSON").Call(
			Qual("net/http", "StatusBadRequest"),
			Map(String()).String().Values(Dict{Lit("err"): Err().Dot("Error").Call()}),
		)),
	)
}

func injectOriginalMethodCall(g *Group, fn *parser.Function, method Code) {
	g.ListFunc(createArgsListFunc(fn.Results, "response")).
		Op("=").Add(method).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
}

func getInitReceiverDepsSignature(fn *parser.Function, info *PackageInfo) (code Code) {
	if fn == nil {
		return
	}
	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			t := field.Type
			if matchFuncType.MatchString(t) {
				return
			}
			depVarName := getReceiverVarName(t)
			g.Id(depVarName).Op("*").Qual(info.Service.Name, strings.Trim(t, "*"))
		}
	})
}

func createTypeFromArgs(name string, args []parser.Field, info *PackageInfo) Code {
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, arg := range args {
			name := arg.Name
			if isArgNameAreDTO(name) {
				name = ""
			}
			field := Id(strings.Title(name)).Op(arg.Prefix)
			if arg.Package != "" {
				field.Qual(info.Service.Name, arg.Type)
			} else {
				field.Id(arg.Type)
			}
			jsonTag := strings.ToLower(name)
			if arg.Type == "error" {
				jsonTag = "-"
			}
			field.Tag(map[string]string{"json": jsonTag})
			g.Add(field)
		}
	})
}

func createReqRespTypes(postfix string, info *PackageInfo) Code {
	code := Comment(fmt.Sprintf("Request/Response types (%s)", postfix)).Line()

	forEachFunction(info.Functions, true, func(fn *parser.Function) {
		_, reqName, respName := getMethodTypes(fn, postfix)
		code.Add(createTypeFromArgs(reqName, fn.Arguments, info))
		code.Line()
		code.Add(createTypeFromArgs(respName, fn.Results, info))
		code.Line()
	})
	return code
}

func createArgsListFunc(args []parser.Field, params ...string) func(*Group) {
	return createArgsList(args, func(arg *Statement) *Statement {
		return arg
	}, params...)
}

func createArgsList(args []parser.Field, transform func(*Statement) *Statement, params ...string) func(*Group) {
	prefix, typeNames := "", ""
	if len(params) > 0 {
		prefix = params[0]
	}
	if len(params) > 1 {
		typeNames = params[1]
	}
	return func(g *Group) {
		for _, arg := range args {
			//Skip iteration if argument type not specified
			if typeNames != "" && !strings.Contains(typeNames, arg.Type+",") {
				continue
			}
			if isArgNameAreDTO(arg.Name) && prefix != "" {
				g.Add(transform(Id(prefix).Dot(arg.Type)))
				return
			}
			name := strings.Title(arg.Name)
			if prefix != "" {
				g.Add(transform(Id(prefix).Dot(name)))
			} else {
				g.Add(transform(Id(name)))
			}
		}
	}
}

func getInitReceiverDepsNames(fn *parser.Function) (code Code) {
	if fn == nil {
		return
	}
	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			injectReceiverName(g, field)
		}
	})
}

var matchFuncType = regexp.MustCompile("^func.*")

//injectReceiverName injects recevier variable name to given scope.
func injectReceiverName(g *Group, field parser.Field) {
	t := field.Type
	if matchFuncType.MatchString(t) {
		return
	}
	depVarName := getReceiverVarName(t)
	g.Id(depVarName)
}

const getEnvHelper = "getEnvHelper"
const firstNotEmptyStrHelper = "firstNotEmptyStrHelper"

func makeHelpers(info *PackageInfo, main *Group, f *File) {
	f.Func().Id(getEnvHelper).Params(Id("envName").String()).String().Block(
		Return(Qual("os", "Getenv").Call(Id("envName"))),
	)

	f.Func().Id(firstNotEmptyStrHelper).Params(Id("a"), Id("b").String()).String().Block(
		If(Id("a").Op("!=").Lit("")).Block(Return(Id("a"))),
		Return(Id("b")),
	)
}
