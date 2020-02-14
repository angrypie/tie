package template

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

const echoPath = "github.com/labstack/echo"
const echoMiddleware = "github.com/labstack/echo/middleware"
const httpModuleId = "HTTP"

func GetServerMainHTTP(info *PackageInfo) (string, error) {
	f := NewFile(strings.ToLower(httpModuleId))

	f.Func().Id("Main").Params().BlockFunc(func(main *Group) {
		makeGracefulShutdown(info, main, f)
		makeInitService(info, main)

		makeStartHTTPServer(info, main, f)
	})
	makeHandlers(info, f, makeHTTPHandler)
	f.Add(createReqRespTypes(httpModuleId, info))
	makeHelpersHTTP(f)

	return fmt.Sprintf("%#v", f), nil
}

func makeHTTPHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	_, request, response := getMethodTypes(fn, httpModuleId)
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

		middlewares := middlewaresMap{
			"getEnv":    Id(getEnvHelper),
			"getHeader": Id(getHeaderHelper).Call(Id("ctx")),
		}
		makeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrHTTP)

		g.Return(Id("ctx").Dot("JSON").Call(Qual("net/http", "StatusOK"), Id("response")))
	}

	makeHandlerWrapper(
		httpModuleId, handlerBody, info, fn, file,
		Id("ctx").Qual(echoPath, "Context"),
		Err().Error(),
	)

}

func makeStartHTTPServer(info *PackageInfo, main *Group, f *File) {
	main.Go().Id("startServer").Call()

	f.Func().Id("startServer").Params().BlockFunc(func(g *Group) {
		makeStartServerInit(info, g)      //SIM
		makeReceiversForHandlers(info, g) //SIM

		g.Id("server").Op(":=").Qual(echoPath, "New").Call()

		//.2 Add handler for each function.
		forEachFunction(info, true, func(fn *parser.Function) {
			handler, _, _ := getMethodTypes(fn, httpModuleId)
			route := fmt.Sprintf("%s/%s", fn.Receiver.Type, fn.Name)

			g.Id("server").Dot("POST").Call(
				Lit(toSnakeCase(route)),
				Id(handler).CallFunc(makeHandlerWrapperCall(fn, info, func(depName string) Code {
					return Id(depName)
				})),
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
