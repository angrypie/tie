package rpcmod

import (
	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

func NewClientModule(p *parser.Parser) template.Module {
	return template.NewStandartModule("client", GenerateClient, p, nil)
}

func GenerateClient(p *parser.Parser) (pkg *template.Package) {
	pkg = &template.Package{}
	return
}

func NewUpgradedModule(p *parser.Parser) template.Module {
	return template.NewStandartModule("client", GenerateUpgraded, p, nil)
}

func GenerateUpgraded(p *parser.Parser) (pkg *template.Package) {
	//p.UpgradeApiImports(imports)
	//files := p.ToFiles()
	pkg = &template.Package{}
	return
}
