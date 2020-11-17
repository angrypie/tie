package dapr

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	"github.com/angrypie/tie/template/modutils"
	"github.com/angrypie/tie/template/protobuf"
	. "github.com/dave/jennifer/jen"
)

const daprModuleId = "DaprIo"
const daprClient = "github.com/dapr/go-sdk/client"
const daprCommon = "github.com/dapr/go-sdk/service/common"
const daprService = "github.com/dapr/go-sdk/service/grpc"
const json = "encoding/json"

//TODO
func GenerateClient(p *parser.Parser) (pkg *template.Package) {
	info := template.NewPackageInfoFromParser(p)
	//TODO all modules needs to create upgraded subpackage to make ServicePath reusable,
	info.SetServicePath(info.Service.Name + "/tie_modules/daprmod/upgraded")
	f := NewFile(strings.ToLower(daprModuleId))

	template.TemplateClient(info, f, func(ids template.ClientMethodIds, g *Group) {
		id := template.ID
		client, data, content, out := id("client"), id("data"), id("content"), id("out")

		//TODO maybe use global dapr client declaration
		g.List(Id(client), Id(ids.Err)).Op(":=").Qual(daprClient, "NewClient").Call()
		template.AddIfErrorGuard(g, nil, ids.Err, nil)
		g.Defer().Id(client).Dot("Close").Call()

		g.List(Id(data), Id(ids.Err)).Op(":=").Qual(json, "Marshal").Call(Id(ids.Request))
		template.AddIfErrorGuard(g, nil, ids.Err, nil)

		g.Id(content).Op(":=").Op("&").Qual(daprClient, "DataContent").
			Values(Dict{
				Id("ContentType"): Lit("application/json"),
				Id("Data"):        Id(data),
			})

		g.List(Id(out), Id(ids.Err)).Op(":=").
			Id(client).Dot("InvokeServiceWithContent").Call(
			Qual("context", "TODO").Call(), Lit(ids.Resource), Lit(ids.Method), Id(content))
		template.AddIfErrorGuard(g, nil, ids.Err, nil)

		g.Err().Op("=").Qual(json, "Unmarshal").
			Call(Id(out), Id(ids.Response))
	})

	return modutils.NewPackage("client", "client.go", f.GoString())
}

func NewClientModule(p *parser.Parser) template.Module {
	return modutils.NewStandartModule("client", GenerateClient, p, nil)
}

func NewModule(p *parser.Parser, services []string) template.Module {
	if p.GetPackageName() == "main" {
		return NewUpgradedModule(p, services)
	}

	deps := []template.Module{
		NewClientModule(p),
		NewUpgradedModule(p, services),
		protobuf.NewModule(p),
	}
	return modutils.NewStandartModule("daprmod", GenerateServer, p, deps)
}

func GenerateUpgraded(p *parser.Parser, services []string) (pkg *template.Package) {
	p.UpgradeApiImports(services, func(i string) (n string) {
		return i + "/tie_modules/daprmod/client"
	})
	files := []modutils.File{}
	for _, file := range p.ToFiles() {
		files = append(files, modutils.File{
			Name:    file.Name,
			Content: file.Content,
		})
	}
	pkg = &template.Package{Name: "upgraded", Files: files}
	return
}

func NewUpgradedModule(p *parser.Parser, services []string) template.Module {
	gen := func(p *parser.Parser) *template.Package {
		return GenerateUpgraded(p, services)
	}
	return modutils.NewStandartModule("upgraded", gen, p, nil)
}

func GenerateServer(p *parser.Parser) *template.Package {
	info := template.NewPackageInfoFromParser(p)
	info.SetServicePath(info.Service.Name + "/tie_modules/daprmod/upgraded")
	f := NewFile(strings.ToLower(daprModuleId))

	template.TemplateRpcServer(info, f, template.TemplateServerConfig{
		GenResourceScope: func(g *Group, resource, instance string) {
			makeStartServer(info, g, f, instance)
		},
		GenHandler: genDaprHandler,
	})

	return modutils.NewPackage("daprmod", "server.go", f.GoString())
}

func genDaprHandler(info *template.PackageInfo, file *File, fn parser.Function) {
	_, request, response := template.GetMethodTypes(fn)
	body := func(g *Group, resourceInstance string) {
		middlewares := template.MiddlewaresMap{"getEnv": Id(template.GetEnvHelper)}
		if len(fn.Arguments) != 0 {
			g.Id("request").Op(":=").New(Id(request))

			stmt := Err().Op("=").Qual(json, "Unmarshal").
				Call(Id("in").Dot("Data"), Id("request"))
			template.AddIfErrorGuard(g, stmt, "err", nil)
		}

		g.Var().Id("response").Id(response)

		template.MakeOriginalCall(
			info, fn, g, middlewares,
			ifDaprHandlerError,
			resourceInstance,
		)

		g.Id("out").Op("=").Op("&").Qual(daprCommon, "Content").
			Values(Dict{Id("ContentType"): Lit("application/json")})

		stmt := List(Id("out").Dot("Data"), Err()).Op("=").
			Qual(json, "Marshal").Call(Id("response"))
		template.AddIfErrorGuard(g, stmt, "err", nil)

		g.Return()
	}

	args := List(
		Id("ctx").Qual("context", "Context"),
		Id("in").Op("*").Qual(daprCommon, "InvocationEvent"),
	)
	resp := List(
		Id("out").Op("*").Qual(daprCommon, "Content"),
		Err().Error(),
	)

	template.MakeHandlerWrapper(file, body, info, fn, args, resp)
}

func ifDaprHandlerError(scope *Group, statement *Statement) {
	template.AddIfErrorGuard(scope, statement, "err", nil)
}

func makeStartServer(info *template.PackageInfo, g *Group, f *File, resourceInstance string) {
	const serverInstance = "DaprService"
	//Init Server
	g.List(Id(serverInstance), Err()).Op(":=").Qual(daprService, "NewService").Call(Lit(":50001"))
	template.AddIfErrorGuard(g, nil, "err", Err())

	//.2 Add handler for each function.
	template.ForEachFunction(info, true, func(fn parser.Function) {
		handler, _, _ := template.GetMethodTypes(fn)

		g.Err().Op("=").Id(serverInstance).Dot("AddServiceInvocationHandler").Call(
			Lit(handler),
			Id(handler).Call(Id(resourceInstance)),
		)
		template.AddIfErrorGuard(g, nil, "err", Err())
	})

	startStmt := Err().Op(":=").Id(serverInstance).Dot("Start").Call()
	template.AddIfErrorGuard(g, startStmt, "err", Err())
}

func getGrpcMethodRoute(fn parser.Function) string {
	route := fmt.Sprintf("%s", fn.Name)
	if fn.Receiver.IsDefined() {
		route = fmt.Sprintf("%s_%s", fn.Receiver.TypeName(), fn.Name)
	}
	return strings.ToLower(route)
}
