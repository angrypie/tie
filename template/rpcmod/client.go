package rpcmod

import (
	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

func NewClientModule(p *parser.Parser) template.Module {
	return template.NewStandartModule("client", GenerateClient, p, nil)
}

func GenerateClient(p *parser.Parser) (pkg *template.Package) {
	pkg = &template.Package{Name: "client", Files: [][]byte{[]byte(`package client`)}}
	return
}

