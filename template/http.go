package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

const echoPath = "github.com/labstack/echo"

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
	forEachFunction(info, true, func(fn *parser.Function) {
		makeHTTPHandler(info, fn, f.Group)
	})
}

func makeHTTPHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	handler, request, response := getMethodTypes(fn, "HTTP")
	receiverVarName := getReceiverVarName(fn.Receiver.Type)
	handlerBody := func(g *Group) {
		//Bind request params
		//Empty argument needs to avoid errors if no other arguments exist
		arguments := createCombinedHandlerArgs(fn, info)
		if len(arguments) != 0 {
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
		}

		//Create response object
		g.Id("response").Op(":=").New(Id(response))

		//If method has receiver generate receiver middleware code
		//else just call public package method
		if hasReceiver(fn) {
			constructorFunc := info.GetConstructor(fn.Receiver.Type)
			if constructorFunc != nil && !hasTopLevelReceiver(constructorFunc, info) {
				receiverType := fn.Receiver.Type
				g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, strings.Trim(receiverType, "*")).Block()
				makeReceiverMiddleware(receiverVarName, g, constructorFunc, info)
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
			constructorFunc := info.GetConstructor(fn.Receiver.Type)
			if constructorFunc == nil || hasTopLevelReceiver(constructorFunc, info) {
				g.Id(receiverVarName).Op("*").Qual(info.Service.Name, strings.Trim(fn.Receiver.Type, "*"))
			} else {
				g.Add(getConstructorDepsSignature(constructorFunc, info))
			}
		}).Params(
			Func().Params(Qual(echoPath, "Context")).Params(Error()),
		).Block(Return(Func().
			Params(Id("ctx").Qual(echoPath, "Context")).
			Params(Err().Error()).BlockFunc(handlerBody),
		)).Line()
	} else {
		file.Func().Id(handler).
			Params(Id("ctx").Qual(echoPath, "Context")).
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

		g.Id("server").Op(":=").Qual(echoPath, "New").Call()
		g.Id("server").Dot("Use").Call(Qual(echoMiddleware, "CORSWithConfig").Call(
			Qual(echoMiddleware, "CORSConfig").Values(Dict{
				Id("AllowOrigins"): Index().String().Values(Lit("*")),
			}),
		))

		//. Set HTTP handlers and init receivers.
		//.1 Create receivers for handlers
		receiversProcessed := make(map[string]bool)
		createReceivers := func(receiverType string, constructorFunc *parser.Function) {
			receiversProcessed[receiverType] = true
			//Skip not top level receivers.
			if constructorFunc != nil && !hasTopLevelReceiver(constructorFunc, info) {
				return
			}
			receiverVarName := getReceiverVarName(receiverType)
			g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, strings.Trim(receiverType, "*")).Block()
			makeReceiverInitialization(receiverVarName, g, constructorFunc, info)
		}
		//Create receivers for each constructor
		for t, c := range info.Constructors {
			createReceivers(t, c)
		}
		//Create receivers that does not have constructor
		forEachFunction(info, false, func(fn *parser.Function) {
			receiverType := fn.Receiver.Type
			//Skip function if it does not have receiver or receiver already created.
			if !hasReceiver(fn) || receiversProcessed[receiverType] {
				return
			}
			constructorFunc := info.GetConstructor(fn.Receiver.Type)
			createReceivers(receiverType, constructorFunc)
		})

		//.2 Add http handler for each function.
		//Route format is /receiver_name/method_name
		forEachFunction(info, true, func(fn *parser.Function) {
			handler, _, _ := getMethodTypes(fn, "HTTP")

			//If handler has receiver
			if hasReceiver(fn) {
				constructorFunc := info.GetConstructor(fn.Receiver.Type)
				receiverType := fn.Receiver.Type
				receiverVarName := getReceiverVarName(receiverType)
				route := fmt.Sprintf("%s/%s", receiverType, fn.Name)

				g.Id("server").Dot("POST").Call(
					Lit(toSnakeCase(route)),
					Id(handler).CallFunc(func(g *Group) {
						if constructorFunc == nil || hasTopLevelReceiver(constructorFunc, info) {
							//Inject receiver to http handler.
							g.Id(receiverVarName)
						} else {
							//Inject dependencies to http handler for non top level receiver.
							g.Add(getConstructorDepsNames(constructorFunc, info))
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
				Func().Params(Id("key").String(), Id("ctx").Qual(echoPath, "Context")).Params(Bool(), Error()).Block(
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

func makeReceiverMiddleware(recId string, scope *Group, constructor *parser.Function, info *PackageInfo) {
	if constructor == nil {
		return
	}
	constructorCall := func(g *Group) {
		for _, field := range constructor.Arguments {
			name := field.Name
			//TODO check getHeader and getEnv function signature
			//Inject getHeader function that returns header of current request
			if name == "getHeader" {
				g.Id(getHeaderHelper).Call(Id("ctx"))
				continue
			}
			//Inject getEnv function that provide access to environment variables
			if name == "getEnv" {
				g.Id(getEnvHelper)
				continue
			}

			//TODO send nil for pointer or empty object otherwise
			if !info.IsReceiverType(field.Type) {
				//g.Id("request").Dot(field.Name)
				g.ListFunc(createArgsListFunc([]parser.Field{field}, "request"))
				continue
			}

			//Oterwise inject receiver dependencie
			g.Id(getReceiverVarName(field.Type))
		}
	}

	ifErrorReturnBadRequestWithErr(
		scope,
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, constructor.Name).CallFunc(constructorCall),
	)
}

func makeReceiverInitialization(recId string, scope *Group, constructor *parser.Function, info *PackageInfo) {
	if constructor == nil {
		return
	}
	constructorCall := func(g *Group) {
		for _, field := range constructor.Arguments {
			name := field.Name
			//TODO check getEnv function signature
			//Inject getEnv function that provide access to environment variables
			if name == "getEnv" {
				g.Id(getEnvHelper)
				continue
			}

			//TODO send nil for pointer or empty object otherwise
			if !info.IsReceiverType(field.Type) {
				g.Nil()
				continue
			}

		}
	}

	scope.If(
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, constructor.Name).CallFunc(constructorCall),
		Err().Op("!=").Nil(),
	).Block(
		//TODO return appropriate error here
		Panic(Err()),
	)

	for _, fn := range info.Functions {
		if fn.Name == "Stop" && info.GetConstructor(fn.Receiver.Type) == constructor {
			scope.Id("stoppableServices").Op("=").Append(Id("stoppableServices"), Id(recId))
			return
		}
	}

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

	forEachFunction(info, true, func(fn *parser.Function) {
		arguments := createCombinedHandlerArgs(fn, info)

		_, reqName, respName := getMethodTypes(fn, postfix)
		code.Add(createTypeFromArgs(reqName, arguments, info))
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

//createArgsList creates list from parser.Field array.
//Transform function are used to modify each element list.
//Optional param 1 is used to specify prefix for each element.
//Optional param 2 is used to specify allowed argument types (format: type1,type2,).
func createArgsList(
	args []parser.Field,
	transform func(*Statement) *Statement,
	params ...string,
) func(*Group) {
	prefix, onlyTypes := "", ""
	if len(params) > 0 {
		prefix = params[0]
	}
	if len(params) > 1 {
		onlyTypes = params[1]
	}
	return func(g *Group) {
		for _, arg := range args {
			//Skip iteration if arg has type that not in onlyTypes (if it is not empty).
			if onlyTypes != "" && !strings.Contains(onlyTypes, arg.Type+",") {
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

var matchFuncType = regexp.MustCompile("^func.*")

func getConstructorDepsNames(fn *parser.Function, info *PackageInfo) (code Code) {
	return getConstructorDeps(fn, info, func(field parser.Field, g *Group) {
		g.Id(getReceiverVarName(field.Type))
	})
}

func getConstructorDepsSignature(fn *parser.Function, info *PackageInfo) (code Code) {
	return getConstructorDeps(fn, info, func(field parser.Field, g *Group) {
		g.Id(getReceiverVarName(field.Type)).Op("*").Qual(info.Service.Name, strings.Trim(field.Type, "*"))
	})
}

func getConstructorDeps(
	fn *parser.Function,
	info *PackageInfo,
	createDep func(field parser.Field, g *Group),
) (code Code) {
	if fn == nil {
		return
	}

	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			t := field.Type
			if matchFuncType.MatchString(t) || !info.IsReceiverType(t) {
				continue
			}
			createDep(field, g)
		}
	})
}

const getEnvHelper = "getEnvHelper"
const firstNotEmptyStrHelper = "firstNotEmptyStrHelper"
const getHeaderHelper = "getHeaderHelper"

func makeHelpers(info *PackageInfo, main *Group, f *File) {
	f.Func().Id(getEnvHelper).Params(Id("envName").String()).String().Block(
		Return(Qual("os", "Getenv").Call(Id("envName"))),
	)

	f.Func().Id(firstNotEmptyStrHelper).Params(Id("a"), Id("b").String()).String().Block(
		If(Id("a").Op("!=").Lit("")).Block(Return(Id("a"))),
		Return(Id("b")),
	)

	f.Func().Id(getHeaderHelper).
		Params(Id("ctx").Qual(echoPath, "Context")).
		Func().Params(String()).String().Block(
		Return(
			Func().Params(Id("headerName").String()).String().Block(
				Return(
					Id("ctx").Dot("Request").Call().Dot("Header").Dot("Get").Call(Id("headerName")),
				),
			),
		),
	)
}
