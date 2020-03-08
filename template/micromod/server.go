package micromod

import (
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	. "github.com/dave/jennifer/jen"
)

const gomicro = "github.com/micro/go-micro/v2"
const gomicroClient = "github.com/micro/go-micro/v2/client"
const microModuleId = "GoMicro"

type PackageInfo = template.PackageInfo

func NewModule(p *parser.Parser, services []string) template.Module {
	if p.GetPackageName() == "main" {
		return NewUpgradedModule(p, services)
	}

	deps := []template.Module{
		NewClientModule(p),
		NewUpgradedModule(p, services),
	}
	return template.NewStandartModule("micromod", GenerateServer, p, deps)
}

func NewUpgradedModule(p *parser.Parser, services []string) template.Module {
	gen := func(p *parser.Parser) *template.Package {
		return GenerateUpgraded(p, services)
	}
	return template.NewStandartModule("upgraded", gen, p, nil)
}

func GenerateUpgraded(p *parser.Parser, services []string) (pkg *template.Package) {
	p.UpgradeApiImports(services, func(i string) (n string) {
		return i + "/tie_modules/micromod/client"
	})
	files := p.ToFiles()
	pkg = &template.Package{Name: "upgraded", Files: files}
	return
}

func GenerateServer(p *parser.Parser) *template.Package {
	info := template.NewPackageInfoFromParser(p)
	info.ServicePath = info.Service.Name + "/tie_modules/micromod/upgraded"
	f := NewFile(strings.ToLower(microModuleId))

	f.Func().Id("Main").Params().BlockFunc(func(main *Group) {
		template.MakeGracefulShutdown(info, main, f)
		template.MakeInitService(info, main)

		makeStartRPCServer(info, main, f)
	})

	template.MakeHandlers(info, f, makeRPCHandler)
	f.Add(template.CreateReqRespTypes(microModuleId, info))
	template.AddGetEnvHelper(f)

	return &template.Package{
		Name:  "micromod",
		Files: [][]byte{[]byte(f.GoString())},
	}
}

func getRpcHandlerArgsList(request, response string) *Statement {
	return List(
		Id("ctx").Qual("context", "Context"),
		Id("request").Op("*").Id(request),
		Id("response").Op("*").Id(response),
	)
}

func makeRPCHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	handlerBody := func(g *Group) {
		middlewares := template.MiddlewaresMap{"getEnv": Id(template.GetEnvHelper)}
		template.MakeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrRPC)
		g.Return(Nil())
	}

	_, request, response := template.GetMethodTypes(fn, microModuleId)

	template.MakeHandlerWrapper(
		microModuleId, handlerBody, info, fn, file,
		getRpcHandlerArgsList(request, response),
		Err().Error(),
	)
}

func getResourceName(info *PackageInfo) string {
	return "Resource__" + info.PackageName
}

func makeStartRPCServer(info *PackageInfo, main *Group, f *File) {
	main.Go().Id("startServer").Call()

	f.Func().Id("startServer").Params().BlockFunc(func(g *Group) {
		receiversCreated := template.MakeReceiversForHandlers(info, g)

		resourceName := getResourceName(info)
		resourceInstance := "Instance__" + resourceName

		f.Type().Id(resourceName).StructFunc(func(g *Group) {
			for receiverType := range receiversCreated {
				receiverVarName := template.GetReceiverVarName(receiverType)
				g.Id(receiverVarName).Op("*").Qual(info.GetServicePath(), template.TrimPrefix(receiverType))
			}
		})

		//.2 Add handler for each function.
		template.ForEachFunction(info, true, func(fn *parser.Function) {
			handler, request, response := template.GetMethodTypes(fn, microModuleId)

			f.Func().Params(Id("resource").Id(resourceName)).Id(handler).
				Params(getRpcHandlerArgsList(request, response)).
				Params(Err().Error()).Block(
				Return(
					Id(handler).
						CallFunc(template.MakeHandlerWrapperCall(fn, info, func(depName string) Code {
							return Id("resource").Dot(depName)
						})).Call(Id("ctx"), Id("request"), Id("response"))))
		})

		g.Id("service").Op(":=").Qual(gomicro, "NewService").Call(
			Qual(gomicro, "Name").Call(Lit(resourceName)),
		)

		g.Id("service").Dot("Init").Call()

		g.Id(resourceInstance).Op(":=").Op("&").Id(resourceName).Values(DictFunc(func(d Dict) {
			for receiverType := range receiversCreated {
				receiverVarName := template.GetReceiverVarName(receiverType)
				d[Id(receiverVarName)] = Id(receiverVarName)
			}
		}))

		g.Qual(gomicro, "RegisterHandler").Call(Id("service").Dot("Server").Call(), Id(resourceInstance))
		g.Id("service").Dot("Run").Call()

	})
}

func ifErrorReturnErrRPC(scope *Group, statement *Statement) {
	template.AddIfErrorGuard(scope, statement, Err())
}
