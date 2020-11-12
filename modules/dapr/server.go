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

const microModuleId = "DaprIo"
const daprCommon = "github.com/dapr/go-sdk/dapr/proto/common/v1"
const daprRuntime = "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
const pbAny = "github.com/golang/protobuf/ptypes/any"
const pbEmpty = "github.com/golang/protobuf/ptypes/empty"
const daprdImport = "github.com/dapr/go-sdk/service/grpc"

//TODO
func GenerateClient(p *parser.Parser) (pkg *template.Package) {
	return
}

func NewClientModule(p *parser.Parser) template.Module {
	return modutils.NewStandartModule("client", GenerateClient, p, nil)
}

func NewModule(p *parser.Parser, services []string) template.Module {
	if p.GetPackageName() == "main" {
		return NewUpgradedModule(p, services)
	}

	deps := []template.Module{
		//NewClientModule(p),
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
	f := NewFile(strings.ToLower(microModuleId))

	template.TemplateRpcServer(info, f, template.TemplateServerConfig{
		GenResourceScope: func(g *Group, resource, instance string) {
			genMethodHandlers(info, g, f, instance)
		},
	})

	return modutils.NewPackage("daprmod", "server.go", f.GoString())
}

func genMethodHandlers(info *template.PackageInfo, g *Group, f *File, resourceInstance string) {
	const serverInstance = "DaprService"
	//Init Server
	g.List(Id(serverInstance), Err()).Op(":=").Qual(daprdImport, "NewService").Call(Lit(":50001"))
	template.AddIfErrorGuard(g, nil, "err", Err())

	startStmt := Err().Op(":=").Id(serverInstance).Dot("Start").Call()
	template.AddIfErrorGuard(g, startStmt, "err", Err())
	//.2 Add handler for each function.
	template.ForEachFunction(info, true, func(fn parser.Function) {
		handler, _, _ := template.GetMethodTypes(fn)

		route := fmt.Sprintf("/%s", fn.Name)
		if fn.Receiver.IsDefined() {
			route = fmt.Sprintf("%s/%s", fn.Receiver.TypeName(), fn.Name)
		}

		g.Id(serverInstance).Dot("AddServiceInvocationHandler").Call(
			Lit(route),
			Id(handler).Call(Id(resourceInstance)),
		)
	})
}
