package httpmod

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	"github.com/angrypie/tie/template/modutils"
	. "github.com/dave/jennifer/jen"
)

const echoPath = "github.com/labstack/echo"
const echoMiddleware = "github.com/labstack/echo/middleware"
const httpModuleId = "HTTP"

type PackageInfo = template.PackageInfo

func NewModule(p *parser.Parser) template.Module {
	return modutils.NewStandartModule("httpmod", GenerateServer, p, nil)
}

func GenerateServer(p *parser.Parser) *template.Package {
	info := template.NewPackageInfoFromParser(p)
	//info.SetServicePath(info.Service.Name + "/tie_modules/httpmod/upgraded")
	f := NewFile(strings.ToLower(httpModuleId))

	template.TemplateRpcServer(info, f, template.TemplateServerConfig{
		GenResourceScope: func(g *Group, resource, instance string) {
			makeStartHTTPServer(info, g, f, instance)
		},
		GenHandler: makeHTTPHandler,
	})

	makeHelpersHTTP(f)

	return modutils.NewPackage("httpmod", "server.go", f.GoString())
}

func makeHTTPHandler(info *PackageInfo, file *File, fn parser.Function) {
	_, request, response := template.GetMethodTypes(fn)
	handlerBody := func(g *Group, resourceInstance string) {
		//Bind request params
		//Empty argument needs to avoid errors if no other arguments exist
		g.Comment("makeHttpHandler body:").Line()
		//TODO is argumens variable unused?
		arguments := template.CreateCombinedHandlerArgs(fn, info)
		if len(arguments) != 0 {
			g.Id("request").Op(":=").New(Id(request))
			g.List(Id("_"), ListFunc(template.CreateArgsListFunc(fn.Arguments, "request", "string,"))).Op("=").
				List(Lit(0), ListFunc(template.CreateArgsList(fn.Arguments, func(arg *Statement, field parser.Field) *Statement {
					return Id(firstNotEmptyStrHelper).Call(
						Id("request").Dot(strings.Title(field.Name())),
						Id("ctx").Dot("QueryParam").Call(Lit(strings.ToLower(arg.GoString()))),
					)
				}, "", "string,")))
			template.AddIfErrorGuard(g, Err().Op(":=").Id("ctx").Dot("Bind").Call(Id("request")), "err", Err())
		}

		//Create response object
		g.Id("response").Op(":=").New(Id(response))

		middlewares := template.MiddlewaresMap{
			"getEnv":    Id(template.GetEnvHelper),
			"getHeader": Id(getHeaderHelper).Call(Id("ctx")),
		}

		template.MakeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrHTTP, resourceInstance)

		g.Return(Id("ctx").Dot("JSON").Call(Qual("net/http", "StatusOK"), Id("response")))
	}

	template.MakeHandlerWrapper(
		file, handlerBody, info, fn,
		Id("ctx").Qual(echoPath, "Context"),
		Err().Error(),
	)
}

func makeStartHTTPServer(info *PackageInfo, g *Group, f *File, resourceInstance string) {

	//generate port variable initialization
	template.MakeStartServerInit(info, g)
	//Create labstack/echo server
	g.Id("server").Op(":=").Qual(echoPath, "New").Call()

	//Add handler for each function.
	template.ForEachFunction(info, true, func(fn parser.Function) {
		handler, _, _ := template.GetMethodTypes(fn)

		route := fmt.Sprintf("/%s", fn.Name)
		if fn.Receiver.IsDefined() {
			route = fmt.Sprintf("/%s/%s", fn.Receiver.TypeName(), fn.Name)
		}

		g.Id("server").Dot("POST").Call(
			Lit(toSnakeCase(route)),
			Id(handler).Call(Id(resourceInstance)),
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
}

func ifErrorReturnErrHTTP(scope *Group, statement *Statement) {
	ret := Id("ctx").Dot("JSON").Call(
		Qual("net/http", "StatusBadRequest"),
		Map(String()).String().Values(Dict{Lit("err"): Err().Dot("Error").Call()}),
	)
	template.AddIfErrorGuard(scope, statement, "err", ret)
}

const firstNotEmptyStrHelper = "firstNotEmptyStrHelper"
const getHeaderHelper = "getHeaderHelper"

func makeHelpersHTTP(f *File) {
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
