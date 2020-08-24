package modutils

import "github.com/angrypie/tie/parser"

type Module interface {
	Name() string
	Generate() *Package
	Deps() []Module
}

type File struct {
	Name    string
	Content []byte
}

type Package struct {
	Name  string
	Files []File
}

func NewPackage(name, fileName, fileContent string) *Package {
	return &Package{
		Name: name,
		Files: []File{{
			Name:    fileName,
			Content: []byte(fileContent),
		}},
	}
}

func TraverseModules(module Module, path []string, cb func(p Module, path []string) error) (err error) {
	err = cb(module, path)
	if err != nil {
		return err
	}

	for _, dep := range module.Deps() {
		err = TraverseModules(dep, append(path, module.Name()), cb)
		if err != nil {
			return err
		}
	}

	return
}

type StandartModule struct {
	name     string
	Parser   *parser.Parser
	deps     []Module
	generate Generator
}

type Generator = func(*parser.Parser) *Package

func NewStandartModule(name string, gen Generator, p *parser.Parser, deps []Module) *StandartModule {
	return &StandartModule{
		name:     name,
		Parser:   p,
		deps:     deps,
		generate: gen,
	}
}

func (module StandartModule) Name() string {
	return module.name
}

func (module StandartModule) Deps() []Module {
	return module.deps
}

func (module StandartModule) Generate() (pkg *Package) {
	if module.generate == nil {
		return
	}
	return module.generate(module.Parser)
}
