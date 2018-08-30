package upgrade

import (
	"bytes"
	"errors"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
)

//Upgrader hold parsed package and uses templates to contruct new, upgraded, packages.
type Upgrader struct {
	Client        bytes.Buffer
	Server        bytes.Buffer
	Service       bytes.Buffer
	Pkg           string
	Parser        *parser.Parser
	ServiceConfig *types.Service
}

//NewUpgrader returns initialized Upgrader
func NewUpgrader(service types.Service) *Upgrader {
	return &Upgrader{
		Pkg:           service.Name,
		ServiceConfig: &service,
		Parser:        parser.NewParser(&service),
	}
}

//Parse parses package and creates various structures for for fourther usage in templates
func (upgrader *Upgrader) Parse() (err error) {
	return upgrader.Parser.Parse(upgrader.Pkg)
}

//Replace replaces each given import with RPC client import
func (upgrader *Upgrader) Replace(imports []string) error {
	ok := upgrader.Parser.UpgradeApiImports(imports)
	if !ok {
		return errors.New("Import deleted but not added")
	}
	return nil
}

//Make builds client, server, service packages to buffers using tempaltes
func (upgrader *Upgrader) Make() (err error) {
	p := upgrader.Parser
	functions, err := p.GetFunctions()
	if err != nil {
		return err
	}

	err = upgrader.initServerUpgrade(p)
	if err != nil {
		return err
	}

	for _, function := range functions {
		//Ignore StopService methods
		if name := function.Name; name == "StopService" {
			continue
		}
		//Contruct both client API lib and API server
		err = upgrader.addApiEndpoint(function)
		if err != nil {
			return err
		}
	}

	err = upgrader.addServerMain(p, functions)
	if err != nil {
		return err
	}

	return err
}

//Upgrade consequentialy calls Replace, Make and Write method
func (upgrader *Upgrader) Upgrade(imports []string) error {
	err := upgrader.Parse()
	if err != nil {
		return err
	}

	err = upgrader.Replace(imports)
	if err != nil {
		return err
	}

	err = upgrader.Make()
	if err != nil {
		return err
	}

	err = upgrader.Write()
	if err != nil {
		return err
	}
	return nil
}
