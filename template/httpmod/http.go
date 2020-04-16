package httpmod

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	. "github.com/dave/jennifer/jen"
)

const echoPath = "github.com/labstack/echo"
const echoMiddleware = "github.com/labstack/echo/middleware"
const httpModuleId = "HTTP"

type PackageInfo = template.PackageInfo

func NewModule(p *parser.Parser) template.Module {
	return template.NewStandartModule("httpmod", GenerateServer, p, nil)
}

func GenerateServer(p *parser.Parser) *template.Package {
	info := template.NewPackageInfoFromParser(p)
	//info.SetServicePath(info.Service.Name + "/tie_modules/httpmod/upgraded")
	f := NewFile(strings.ToLower(httpModuleId))

	f.Func().Id("Main").Params().BlockFunc(func(main *Group) {
		template.MakeGracefulShutdown(info, main, f)
		template.MakeInitService(info, main)

		makeStartHTTPServer(info, main, f)
	})
	template.MakeHandlers(info, f, makeHTTPHandler)
	f.Add(template.CreateReqRespTypes(info))
	makeHelpersHTTP(f)

	return &template.Package{
		Name:  "httpmod",
		Files: [][]byte{[]byte(f.GoString())},
	}
}

func makeHTTPHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	_, request, response := template.GetMethodTypes(fn)
	handlerBody := func(g *Group) {
		//Bind request params
		//Empty argument needs to avoid errors if no other arguments exist
		arguments := template.CreateCombinedHandlerArgs(fn, info)
		if len(arguments) != 0 {
			g.Id("request").Op(":=").New(Id(request))
			g.List(Id("_"), ListFunc(template.CreateArgsListFunc(fn.Arguments, "request", "string,"))).Op("=").
				List(Lit(0), ListFunc(template.CreateArgsList(fn.Arguments, func(arg *Statement, field parser.Field) *Statement {
					return Id(firstNotEmptyStrHelper).Call(
						Id("request").Dot(strings.Title(field.Name)),
						Id("ctx").Dot("QueryParam").Call(Lit(strings.ToLower(arg.GoString()))),
					)
				}, "", "string,")))
			template.AddIfErrorGuard(g, Err().Op(":=").Id("ctx").Dot("Bind").Call(Id("request")), Err())
		}

		//Create response object
		g.Id("response").Op(":=").New(Id(response))

		middlewares := template.MiddlewaresMap{
			"getEnv":    Id(template.GetEnvHelper),
			"getHeader": Id(getHeaderHelper).Call(Id("ctx")),
		}
		template.MakeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrHTTP)

		g.Return(Id("ctx").Dot("JSON").Call(Qual("net/http", "StatusOK"), Id("response")))
	}

	template.MakeHandlerWrapper(
		handlerBody, info, fn, file,
		Id("ctx").Qual(echoPath, "Context"),
		Err().Error(),
	)

}

func makeStartHTTPServer(info *PackageInfo, main *Group, f *File) {
	main.Err().Op(":=").Id("startServer").Call()
	main.If(Err().Op("!=").Nil()).Block(Panic(Err()))

	f.Func().Id("startServer").Params().Params(Err().Error()).BlockFunc(func(g *Group) {
		template.MakeStartServerInit(info, g)
		template.MakeReceiversForHandlers(info, g)

		g.Id("server").Op(":=").Qual(echoPath, "New").Call()

		//.2 Add handler for each function.
		template.ForEachFunction(info, true, func(fn *parser.Function) {
			handler, _, _ := template.GetMethodTypes(fn)

			route := fmt.Sprintf("/%s", fn.Name)
			if fn.Receiver.IsDefined() {
				route = fmt.Sprintf("%s/%s", fn.Receiver.GetLocalTypeName(), fn.Name)
			}

			g.Id("server").Dot("POST").Call(
				Lit(toSnakeCase(route)),
				Id(handler).CallFunc(template.MakeHandlerWrapperCall(fn, info, func(depName string) Code {
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
		g.Return()
	})
}

func ifErrorReturnErrHTTP(scope *Group, statement *Statement) {
	ret := Id("ctx").Dot("JSON").Call(
		Qual("net/http", "StatusBadRequest"),
		Map(String()).String().Values(Dict{Lit("err"): Err().Dot("Error").Call()}),
	)
	template.AddIfErrorGuard(scope, statement, ret)
}

const firstNotEmptyStrHelper = "firstNotEmptyStrHelper"
const getHeaderHelper = "getHeaderHelper"

func makeHelpersHTTP(f *File) {
	template.AddGetEnvHelper(f)

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
					Id(template.GetEnvHelper).Call(Lit("TIE_API_KEY")),
					Lit(key),
				)
				g.Return().List(Id("key").Op("==").Id("auth"), Nil())
			})),
	)
}

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	return strings.ToLower(
		matchAllCap.ReplaceAllString(str, "${1}_${2}"),
	)
}
