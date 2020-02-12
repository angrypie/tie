package template

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

const echoPath = "github.com/labstack/echo"

func GetServerMain(info *PackageInfo) (string, error) {
	f := NewFile("http")

	f.Func().Id("Main").Params().BlockFunc(func(g *Group) {
		makeGracefulShutdown(info, g, f)
		makeInitService(info, g, f)

		makeHTTPServer(info, g, f)
	})

	return fmt.Sprintf("%#v", f), nil
}

func makeHTTPServer(info *PackageInfo, main *Group, f *File) {
	makeStartHTTPServer(info, main, f)
	makeHTTPRequestResponseTypes(info, main, f)
	makeHTTPHandlers(info, main, f)
	makeHelpersHTTP(f)
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
				List(Lit(0), ListFunc(createArgsList(fn.Arguments, func(arg *Statement, field parser.Field) *Statement {
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
				g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, trimPrefix(receiverType)).Block()
				makeReceiverMiddlewareHTTP(receiverVarName, g, constructorFunc, info)
			}
			injectOriginalMethodCall(g, fn, Id(receiverVarName).Dot(fn.Name))
		} else {
			injectOriginalMethodCall(g, fn, Qual(info.Service.Name, fn.Name))
		}

		ifErrorReturnErrHTTP(
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
				g.Id(receiverVarName).Op("*").Qual(info.Service.Name, trimPrefix(fn.Receiver.Type))
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
			g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, trimPrefix(receiverType)).Block()
			makeReceiverInitialization(receiverVarName, g, constructorFunc, info)
		}
		//Create receivers for each constructor
		for t, c := range info.Constructors {
			createReceivers(t, c)
		}
		//TODO
		//Create receivers that does not have constructor
		forEachFunction(info, false, func(fn *parser.Function) {
			receiverType := fn.Receiver.Type
			//Skip function if it does not have receiver or receiver already created.
			if !hasReceiver(fn) || receiversProcessed[receiverType] {
				return
			}
			//It will not create constructor call due constructor func is nil
			createReceivers(receiverType, nil)
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
					Id(handler).
						CallFunc(func(g *Group) {
							if constructorFunc == nil || hasTopLevelReceiver(constructorFunc, info) {
								//Inject receiver to http handler.
								g.Id(receiverVarName)
							} else {
								//Inject dependencies to http handler for non top level receiver.
								g.Add(getConstructorDeps(constructorFunc, info, func(field parser.Field, g *Group) {
									g.Id(getReceiverVarName(field.Type))
								}))
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

func makeReceiverMiddlewareHTTP(recId string, scope *Group, constructor *parser.Function, info *PackageInfo) {
	if constructor == nil {
		return
	}

	constructorCall := makeCallWithMiddleware(constructor, info, middlewaresMap{
		"getEnv":    Id(getEnvHelper),
		"getHeader": Id(getHeaderHelper).Call(Id("ctx")),
	})

	ifErrorReturnErrHTTP(
		scope,
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, constructor.Name).CallFunc(constructorCall),
	)
}

func ifErrorReturnErrHTTP(scope *Group, statement *Statement) {
	ret := Id("ctx").Dot("JSON").Call(
		Qual("net/http", "StatusBadRequest"),
		Map(String()).String().Values(Dict{Lit("err"): Err().Dot("Error").Call()}),
	)
	addIfErrorGuard(scope, statement, ret)
}

const firstNotEmptyStrHelper = "firstNotEmptyStrHelper"
const getHeaderHelper = "getHeaderHelper"

func makeHelpersHTTP(f *File) {
	addGetEnvHelper(f)

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
