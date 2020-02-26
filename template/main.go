package template

import (
	"fmt"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
	. "github.com/dave/jennifer/jen"
)

func GetMainPackage(packagePath string, modules []string) *Package {
	f := NewFile("main")

	f.Func().Id("main").Params().BlockFunc(func(g *Group) {
		for _, module := range modules {
			path := fmt.Sprintf("%s/tie_modules/%s", packagePath, module)
			g.Qual(path, "Main").Call()
		}
		makeWaitGuard(g)
	})

	return &Package{
		Name:  "main",
		Files: [][]byte{[]byte(f.GoString())},
	}
}

func NewMainModule(p *parser.Parser, deps []Module) Module {
	var modules []string
	for _, dep := range deps {
		modules = append(modules, dep.Name())
	}

	generator := func(p *parser.Parser) *Package {
		return GetMainPackage(p.Service.Name, modules)
	}
	return NewStandartModule("tie_modules", generator, p, deps)
}

type PackageInfo struct {
	Functions     []*parser.Function
	Constructors  map[string]*parser.Function
	IsInitService bool
	IsStopService bool
	Service       *types.Service
}

func (info *PackageInfo) IsReceiverType(t string) bool {
	_, ok := info.Constructors[t]
	return ok
}

func (info *PackageInfo) GetConstructor(t string) *parser.Function {
	return info.Constructors[t]
}

func NewPackageInfoFromParser(p *parser.Parser) *PackageInfo {
	functions := p.GetFunctions()

	var fns []*parser.Function
	for _, fn := range functions {
		if name := fn.Name; name == "InitService" || name == "StopService" {
			continue
		}
		fns = append(fns, fn)
	}

	info := PackageInfo{
		Functions:    fns,
		Service:      p.Service,
		Constructors: make(map[string]*parser.Function),
	}

	for _, fn := range functions {
		if fn.Name == "InitService" {
			info.IsInitService = true
		}
		if fn.Name == "StopService" {
			info.IsStopService = true
		}

		if _, ok := info.Constructors[fn.Receiver.Type]; ok {
			continue
		}

		ok, receiverType := isConventionalConstructor(fn)
		if ok {
			info.Constructors[receiverType] = fn
		}
	}

	return &info
}

func createErrLog(msg string) *Statement {
	return Qual("log", "Printf").Call(List(Lit("ERR %s: %s"), Lit(msg), Err()))
}
