package template

import "github.com/angrypie/tie/parser"

type Module interface {
	Name() string
	Generate() *Package
	Deps() []Module
}

type Package struct {
	Name string
	Code string
}

func TraverseModules(module Module, cb func(p Module) error) (err error) {
	err = cb(module)
	if err != nil {
		return err
	}

	for _, dep := range module.Deps() {
		err = TraverseModules(dep, cb)
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

type Generator = func(*parser.Parser) string

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
	pkg = &Package{
		Name: module.Name(),
		Code: module.generate(module.Parser),
	}
	return
}
