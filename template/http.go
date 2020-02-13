package template

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

const echoPath = "github.com/labstack/echo"
const echoMiddleware = "github.com/labstack/echo/middleware"

func GetServerMainHTTP(info *PackageInfo) (string, error) {
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
			addIfErrorGuard(g, Err().Op(":=").Id("ctx").Dot("Bind").Call(Id("request")), Err())
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
	file.Func().Id(handler).ParamsFunc(func(g *Group) {
		if !hasReceiver(fn) {
			return
		}
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

}

func makeHTTPRequestResponseTypes(info *PackageInfo, main *Group, f *File) {
	f.Add(createReqRespTypes("HTTP", info))
}

func makeStartHTTPServer(info *PackageInfo, main *Group, f *File) {
	main.Go().Id("startServer").Call()

	f.Func().Id("startServer").Params().BlockFunc(func(g *Group) {
		makeStartServerInit(info, g)      //SIM
		makeReceiversForHandlers(info, g) //SIM

		g.Id("server").Op(":=").Qual(echoPath, "New").Call()

		//.2 Add handler for each function.
		forEachFunction(info, true, func(fn *parser.Function) {
			handler, _, _ := getMethodTypes(fn, "HTTP")

			route := fmt.Sprintf("%s/%s", fn.Receiver.Type, fn.Name)

			g.Id("server").Dot("POST").Call(
				Lit(toSnakeCase(route)),
				Id(handler).CallFunc(makeHandlerWrapperCall(fn, info)),
			)
		})

		//Configuration before start
		g.Id("server").Dot("Use").Call(Qual(echoMiddleware, "CORSWithConfig").Call(
			Qual(echoMiddleware, "CORSConfig").Values(Dict{
				Id("AllowOrigins"): Index().String().Values(Lit("*")),
			}),
		))
		//Enable authentication if auth field is specified in config
		addAuthenticationHTTP(info, g)
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

func addAuthenticationHTTP(info *PackageInfo, g *Group) {
	key := info.Service.Auth
	if key == "" {
		return
	}
	g.Id("server").Dot("Use").Call(Qual(echoMiddleware, "KeyAuth").Call(
		Func().Params(Id("key").String(), Id("ctx").Qual(echoPath, "Context")).Params(Bool(), Error()).
			BlockFunc(func(g *Group) {
				g.Id("auth").Op(":=").Id(firstNotEmptyStrHelper).Call(
					Id(getEnvHelper).Call(Lit("TIE_API_KEY")),
					Lit(key),
				)
				g.Return().List(Id("key").Op("==").Id("auth"), Nil())
			})),
	)
}
