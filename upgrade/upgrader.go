package upgrade

import (
	"bytes"

	"github.com/angrypie/tie/parser"
)

type Upgrader struct {
	Client  bytes.Buffer
	Server  bytes.Buffer
	Service bytes.Buffer
	Pkg     string
	Parser  *parser.Parser
}

func NewUpgrader(pkgPath string) *Upgrader {
	return &Upgrader{
		Pkg:    pkgPath,
		Parser: parser.NewParser(),
	}
}

func (upgrader *Upgrader) Parse() (err error) {
	return upgrader.Parser.Parse(upgrader.Pkg)
}

func (upgrader *Upgrader) Replace(imports []string) (ok bool) {
	ok = upgrader.Parser.UpgradeApiImports(imports)
	return ok
}

func (upgrader *Upgrader) Make() (err error) {
	p := upgrader.Parser
	functions, err := p.GetFunctions()
	if err != nil {
		return err
	}
	upgrader.initServerUpgrade(p)

	for _, function := range functions {
		//Contruct both client API lib and API server
		err = upgrader.addApiEndpoint(function)
		if err != nil {
			return err
		}
	}

	err = upgrader.addServerMain(p)
	if err != nil {
		return err
	}

	return err
}
