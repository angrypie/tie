package upgrade

import (
	"bytes"
	"errors"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	"github.com/angrypie/tie/types"
)

//Upgrader hold parsed package and uses templates to contruct new, upgraded, packages.
type Upgrader struct {
	//RPC API client package
	Client bytes.Buffer
	//RPC or/and HTTP API server package
	Server map[string]*bytes.Buffer
	//Original package with replaced import.
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
		Server:        make(map[string]*bytes.Buffer),
	}
}

//Upgrade consequentialy calls Parse, Replace, Make and Write method
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

//Parse parses package and creates various structures for for fourther usage in templates
func (upgrader *Upgrader) Parse() (err error) {
	return upgrader.Parser.Parse(upgrader.Pkg)
}

//Replace replaces each given import with RPC client import
func (upgrader *Upgrader) Replace(imports []string) error {
	ok := upgrader.Parser.UpgradeApiImports(imports)
	if !ok {
		return errors.New("import deleted but not added")
	}
	return nil
}

//Make builds client, server, service packages to buffers using tempaltes
func (upgrader *Upgrader) Make() (err error) {
	p := upgrader.Parser
	//TODO create subpackage for each upgrader

	if upgrader.Parser.Service.Type == "httpOnly" {
		info, err := template.NewPackageInfoFromParser(p)
		if err != nil {
			return err
		}
		serverStr, err := template.GetServerMain(info)
		if err != nil {
			return err
		}

		upgrader.Server["http"] = bytes.NewBufferString(serverStr)
	} else {
		functions, err := p.GetFunctions()
		if err != nil {
			return err
		}
		err = upgrader.initServerUpgrade(p)
		if err != nil {
			return err
		}

		for _, function := range functions {
			if name := function.Name; name == "StopService" {
				continue
			}
			err = upgrader.addApiEndpoint(function)
			if err != nil {
				return err
			}
		}
		err = upgrader.addServerMain(p, functions)
		if err != nil {
			return err
		}
	}

	main, err := template.GetMainPackage(upgrader.Parser.Service.Name, upgrader.serverModulesDirs())
	if err != nil {
		return err
	}

	upgrader.Server["main"] = bytes.NewBufferString(main)

	return err
}

func (upgrader Upgrader) serverModulesDirs() (modules []string) {
	for module := range upgrader.Server {
		modules = append(modules, "tie_server_"+module)
	}
	return
}
