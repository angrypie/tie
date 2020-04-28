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
	Functions     []parser.Function
	Constructors  map[string]*TypeConstructor
	PackageName   string
	IsInitService bool
	IsStopService bool
	Service       *types.Service
	//ServicePath should refer to modified original package.
	servicePath string
}

func (info PackageInfo) GetServicePath() string {
	if info.servicePath == "" {
		return info.Service.Name
	}
	return info.servicePath
}

func (info *PackageInfo) SetServicePath(path string) {
	info.servicePath = path
}

func (info PackageInfo) IsReceiverType(field parser.Field) bool {
	return info.GetConstructor(field) != nil
}

func (info PackageInfo) GetConstructor(field parser.Field) *TypeConstructor {
	return info.Constructors[field.GetLocalTypeName()]
}

func NewPackageInfoFromParser(p *parser.Parser) *PackageInfo {
	functions := p.GetFunctions()

	var fns []parser.Function
	for _, fn := range functions {
		if name := fn.Name; name == "InitService" || name == "StopService" {
			continue
		}
		fns = append(fns, fn)
	}

	info := PackageInfo{
		Functions:    fns,
		Service:      p.Service,
		Constructors: make(map[string]*TypeConstructor),
		PackageName:  p.GetPackageName(),
	}

	for _, fn := range functions {
		if fn.Name == "InitService" {
			info.IsInitService = true
		}
		if fn.Name == "StopService" {
			info.IsStopService = true
		}

		receiver, ok := isConventionalConstructor(fn)
		if ok {
			info.Constructors[receiver.GetLocalTypeName()] = NewTypeConstructor(fn, receiver)
		}
	}

	return &info
}

func createErrLog(msg string) *Statement {
	return Qual("log", "Printf").Call(List(Lit("ERR %s: %s"), Lit(msg), Err()))
}

type TypeConstructor struct {
	Function parser.Function
	Receiver parser.Field
}

func NewTypeConstructor(fn parser.Function, rec parser.Field) (constructor *TypeConstructor) {
	return &TypeConstructor{
		Function: fn,
		Receiver: rec,
	}
}
