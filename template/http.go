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
	makeFirstNotEmptyStr(info, main, f)

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
					return Id("firstNotEmptyStr").Call(
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

			if isTopLevelInitReceiver(initReceiverFunc) {
				g.ListFunc(createArgsListFunc(fn.Results, "response")).
					Op("=").Id(receiverVarName).Dot(fn.Name).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
			} else {
				receiverType := fn.Receiver.Type
				g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, strings.Trim(receiverType, "*")).Block()
				makeReceiverMiddleware(receiverVarName, g, initReceiverFunc)
			}

			g.ListFunc(createArgsListFunc(fn.Results, "response")).
				Op("=").Id(receiverVarName).Dot(fn.Name).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
		} else {
			g.ListFunc(createArgsListFunc(fn.Results, "response")).
				Op("=").Qual(info.Service.Name, fn.Name).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
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
			if isTopLevelInitReceiver(initReceiverFunc) {
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
			if !isTopLevelInitReceiver(initReceiverFunc) {
				return
			}
			receiverVarName := getReceiverVarName(receiverType)
			g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, strings.Trim(receiverType, "*")).Block()
			makeReceiverInitialization(receiverVarName, g, initReceiverFunc)
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
						if isTopLevelInitReceiver(initReceiverFunc) {
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
						Id("envKey").Op(":=").Qual("os", "Getenv").Call(Lit("TIE_API_KEY")),
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

func makeFirstNotEmptyStr(info *PackageInfo, main *Group, f *File) {
	f.Func().Id("firstNotEmptyStr").Params(Id("a"), Id("b").String()).String().Block(
		If(Id("a").Op("!=").Lit("")).Block(Return(Id("a"))),
		Return(Id("b")),
	)
}

func forEachFunction(fns []*parser.Function, skipInit bool, cb func(*parser.Function)) {
	for _, fn := range fns {
		//Skip InitReceiver function
		if skipInit && fn.Receiver.Name != "" && fn.Name == "InitReceiver" {
			continue
		}
		cb(fn)
	}
}

func findInitReceiver(fns []*parser.Function, forFunc *parser.Function) *parser.Function {
	for _, fn := range fns {
		if fn.Name != "InitReceiver" {
			continue
		}
		reg := fmt.Sprintf(`\A\*?%s\z`, forFunc.Receiver.Type)
		if match, _ := regexp.MatchString(reg, fn.Receiver.Type); !match {
			continue
		}

		return fn
	}
	return nil
}

func makeReceiverMiddleware(recId string, scope *Group, initReceiver *parser.Function) {
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
				envName := "envName"
				g.Func().Params(Id(envName).String()).String().Block(
					Return(Qual("os", "Getenv").Call(Id(envName))),
				)
				continue
			}

			//Oterwise inject receiver dependencie
			injectReceiverName(g, field)
		}
	}

	ifErrorReturnBadRequestWithErr(
		scope,
		Err().Op(":=").Id(recId).Dot("InitReceiver").CallFunc(initReceiverCall),
	)
}

func makeReceiverInitialization(recId string, scope *Group, initReceiver *parser.Function) {
	if initReceiver == nil {
		return
	}
	initReceiverCall := func(g *Group) {
		for _, field := range initReceiver.Arguments {
			name := field.Name
			//TODO check getEnv function signature
			//Inject getEnv function that provide access to environment variables
			if name == "getEnv" {
				envName := "envName"
				g.Func().Params(Id(envName).String()).String().Block(
					Return(Qual("os", "Getenv").Call(Id(envName))),
				)
			}
		}
	}

	scope.If(
		Err().Op(":=").Id(recId).Dot("InitReceiver").CallFunc(initReceiverCall),
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
