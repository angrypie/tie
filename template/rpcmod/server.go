package rpcmod

import (
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	. "github.com/dave/jennifer/jen"
)

const rpcxServer = "github.com/smallnest/rpcx/server"
const rpcModuleId = "RPC"

type PackageInfo = template.PackageInfo

func NewModule(p *parser.Parser, services []string) template.Module {
	if p.GetPackageName() == "main" {
		return NewUpgradedModule(p, services)
	}

	deps := []template.Module{
		NewClientModule(p),
		NewUpgradedModule(p, services),
	}
	return template.NewStandartModule("rpcmod", GenerateServer, p, deps)
}

func NewUpgradedModule(p *parser.Parser, services []string) template.Module {
	gen := func(p *parser.Parser) *template.Package {
		return GenerateUpgraded(p, services)
	}
	return template.NewStandartModule("upgraded", gen, p, nil)
}

func GenerateUpgraded(p *parser.Parser, services []string) (pkg *template.Package) {
	p.UpgradeApiImports(services, func(i string) (n string) {
		return i + "/tie_modules/rpcmod/client"
	})
	files := p.ToFiles()
	pkg = &template.Package{Name: "upgraded", Files: files}
	return
}

func GenerateServer(p *parser.Parser) *template.Package {
	info := template.NewPackageInfoFromParser(p)
	info.ServicePath = info.Service.Name + "/tie_modules/rpcmod/upgraded"
	f := NewFile(strings.ToLower(rpcModuleId))

	f.Func().Id("Main").Params().BlockFunc(func(main *Group) {
		template.MakeGracefulShutdown(info, main, f)
		template.MakeInitService(info, main)

		makeStartRPCServer(info, main, f)
	})

	template.MakeHandlers(info, f, makeRPCHandler)
	f.Add(template.CreateReqRespTypes(rpcModuleId, info))
	template.AddGetEnvHelper(f)

	return &template.Package{
		Name:  "rpcmod",
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

	_, request, response := template.GetMethodTypes(fn, rpcModuleId)

	template.MakeHandlerWrapper(
		rpcModuleId, handlerBody, info, fn, file,
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
		template.MakeStartServerInit(info, g) //SIM
		receiversCreated := template.MakeReceiversForHandlers(info, g)

		//RC replace http server init
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
			handler, request, response := template.GetMethodTypes(fn, rpcModuleId)

			f.Func().Params(Id("resource").Id(resourceName)).Id(handler).
				Params(getRpcHandlerArgsList(request, response)).
				Params(Err().Error()).Block(
				Return(
					Id(handler).
						CallFunc(template.MakeHandlerWrapperCall(fn, info, func(depName string) Code {
							return Id("resource").Dot(depName)
						})).Call(Id("ctx"), Id("request"), Id("response"))))
		})

		g.Id("server").Op(":=").Qual(rpcxServer, "NewServer").Call()
		g.Id(resourceInstance).Op(":=").Op("&").Id(resourceName).Values(DictFunc(func(d Dict) {
			for receiverType := range receiversCreated {
				receiverVarName := template.GetReceiverVarName(receiverType)
				d[Id(receiverVarName)] = Id(receiverVarName)
			}
		}))

		addMDNSRegistry(g, info)

		g.Id("server").Dot("RegisterName").Call(Lit(resourceName), Id(resourceInstance), Lit(""))
		g.Id("server").Dot("Serve").Call(Lit("tcp"), Id("address"))

		//RC end
	})
}

func addMDNSRegistry(g *Group, info *PackageInfo) {
	resourceName := getResourceName(info)

	g.List(Id("zconfServer"), Id("err")).Op(":=").Qual(zeroconf, "Register").Call(
		Lit("GoZeroconf"), Lit(resourceName),
		Lit("local."), Id("port"),
		Index().Id("string").Values(Lit("txtv=0"), Lit("lo=1"), Lit("la=2")), Id("nil"),
	)
	g.If(Err().Op("!=").Nil()).Block(Panic(Err()))

	g.Defer().Id("zconfServer").Dot("Shutdown").Call()
}

func ifErrorReturnErrRPC(scope *Group, statement *Statement) {
	template.AddIfErrorGuard(scope, statement, Err())
}
