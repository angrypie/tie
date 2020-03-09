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

func makeRPCHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	handlerBody := func(g *Group) {
		middlewares := template.MiddlewaresMap{"getEnv": Id(template.GetEnvHelper)}
		template.MakeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrRPC)
		g.Return(Nil())
	}

	_, request, response := template.GetMethodTypes(fn, rpcModuleId)

	template.MakeHandlerWrapper(
		rpcModuleId, handlerBody, info, fn, file,
		template.GetRpcHandlerArgsList(request, response),
		Err().Error(),
	)
}

func makeStartRPCServer(info *PackageInfo, main *Group, f *File) {
	template.MakeStartRPCServer(info, rpcModuleId, main, f, func(g *Group, resource, instance string) {
		template.MakeStartServerInit(info, g)

		g.Id("server").Op(":=").Qual(rpcxServer, "NewServer").Call()
		addMDNSRegistry(g, info)

		g.Id("server").Dot("RegisterName").Call(Lit(resource), Id(instance), Lit(""))
		g.Id("server").Dot("Serve").Call(Lit("tcp"), Id("address"))
	})

}

func addMDNSRegistry(g *Group, info *PackageInfo) {
	resourceName := template.GetResourceName(info)

	g.List(Id("zconfServer"), Id("err")).Op(":=").Qual(zeroconf, "Register").Call(
		Lit("GoZeroconf"), Lit(resourceName),
		Lit("local."), Id("port"),
		Index().Id("string").Values(Lit("txtv=0"), Lit("lo=1"), Lit("la=2")), Id("nil"),
	)
	template.AddIfErrorGuard(g, nil, nil)

	g.Defer().Id("zconfServer").Dot("Shutdown").Call()
}

func ifErrorReturnErrRPC(scope *Group, statement *Statement) {
	template.AddIfErrorGuard(scope, statement, Err())
}
