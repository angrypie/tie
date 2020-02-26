package upgrade

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	"github.com/angrypie/tie/template/httpmod"
	"github.com/angrypie/tie/template/rpcmod"
	"github.com/angrypie/tie/types"
	"github.com/spf13/afero"
)

//Upgrader hold parsed package and uses templates to contruct new, upgraded, packages.
type Upgrader struct {
	//RPC API client package
	Client bytes.Buffer
	//RPC or/and HTTP API server package
	Module map[string]*bytes.Buffer
	//Original package with replaced import.
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
		Module:        make(map[string]*bytes.Buffer),
	}
}

//Upgrade consequentialy calls Parse, Replace, Make and Write method
func (upgrader *Upgrader) Upgrade(services []string) error {
	err := upgrader.Parse()
	if err != nil {
		return err
	}

	err = upgrader.GenerateModules(services)
	if err != nil {
		return err
	}

	return nil
}

//Parse parses package and creates various structures for for fourther usage in templates.
func (upgrader *Upgrader) Parse() (err error) {
	return upgrader.Parser.Parse(upgrader.Pkg)
}

//GenerateModules genarates modules code.
func (upgrader *Upgrader) GenerateModules(services []string) (err error) {
	p := upgrader.Parser
	servicePath := p.Package.Path

	types := strings.Split(upgrader.Parser.Service.Type, " ")

	var modules []template.Module

	for _, serviceType := range types {
		switch serviceType {
		case "http":
			modules = append(modules, httpmod.NewModule(p))
		case "rpc":
			modules = append(modules, rpcmod.NewModule(p, services))
		default:
			modules = append(modules, rpcmod.NewModule(p, services))
		}
	}

	module := template.NewMainModule(p, modules)

	err = writeHelper(servicePath, "tie_modules/tie_upgraded", upgrader.Parser.ToFiles()...)
	if err != nil {
		return
	}
	template.TraverseModules(module, []string{""},
		func(m template.Module, modulePath []string) (err error) {
			fsPath := path.Join(servicePath, strings.Join(modulePath, "/"))
			pkg := m.Generate()

			return writeHelper(fsPath, m.Name(), pkg.Files...)
		})
	return err
}

//Clean removes files and directories created by Write method
func (upgrader *Upgrader) Clean() error {
	fs := afero.NewOsFs()
	modulesDir := path.Join(upgrader.Parser.Package.Path, "tie_modules")
	return fs.RemoveAll(modulesDir)
}

//writeHelper creates directory for package and write files.
func writeHelper(path, dir string, files ...[]byte) error {
	fs := afero.NewOsFs()
	fullPath := fmt.Sprintf("%s/%s", path, dir)

	err := fs.MkdirAll(fullPath, 0755)
	if err != nil {
		return err
	}

	for index, file := range files {
		err = afero.WriteFile(
			fs,
			fmt.Sprintf("%s/%d.go", fullPath, index),
			file,
			0644,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
