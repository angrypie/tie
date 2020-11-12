package dapr

import (
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	"github.com/angrypie/tie/template/modutils"
	"github.com/angrypie/tie/template/protobuf"
	. "github.com/dave/jennifer/jen"
)

const microModuleId = "DaprIo"

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

	template.TemplateRpcServer(info, f, func(g *Group, resource, instance string) {
		genInitGrpcServer(g, instance)
		genMethodHandlers(info, g, f)
	})

	return modutils.NewPackage("daprmod", "server.go", f.GoString())
}
