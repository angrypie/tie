package template

import (
	"fmt"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
	. "github.com/dave/jennifer/jen"
)

func GetMainPackage(packagePath string, modules []string) (data string, err error) {
	f := NewFile("main")

	f.Func().Id("main").Params().BlockFunc(func(g *Group) {
		for _, module := range modules {
			path := fmt.Sprintf("%s/%s", packagePath, module)
			g.Qual(path, "Main").Call()
		}
		makeWaitGuard(g)
	})

	return fmt.Sprintf("%#v", f), nil
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

func NewPackageInfoFromParser(p *parser.Parser) (*PackageInfo, error) {
	functions, err := p.GetFunctions()
	if err != nil {
		return nil, err
	}
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

	return &info, nil
}

func createErrLog(msg string) *Statement {
	return Qual("log", "Printf").Call(List(Lit("ERR %s: %s"), Lit(msg), Err()))
}
