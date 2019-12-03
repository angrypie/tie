package template

import (
	"fmt"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
	. "github.com/dave/jennifer/jen"
)

func GetServerMain(info *PackageInfo) (string, error) {
	f := NewFile("main")

	f.Func().Id("main").Params().BlockFunc(func(g *Group) {
		makeGracefulShutdown(info, g, f)
		makeInitService(info, g, f)

		makeHTTPServer(info, g, f)

		makeWaitGuard(g)
	})

	return fmt.Sprintf("%#v", f), nil
}

func makeWaitGuard(main *Group) {
	main.Op("<-").Make(Chan().Bool())
}

func makeInitService(info *PackageInfo, main *Group, f *File) {
	if !info.IsInitService {
		return
	}
	main.If(
		Err().Op(":=").Qual(info.Service.Name, "InitService").Call(),
		Err().Op("!=").Nil(),
	).Block(
		createErrLog("failed to init service"),
		Return(),
	)
}

func makeGracefulShutdown(info *PackageInfo, g *Group, f *File) {
	functionName := "gracefulShutDown"
	g.Id(functionName).Call()

	f.Type().Id("stoppable").Interface(Id("Stop").Params().Error())

	f.Var().Id("stoppableServices").Index().Id("stoppable")

	f.Func().Id(functionName).Params().Block(
		Id("sigChan").Op(":=").Make(Chan().Qual("os", "Signal")),
		Qual("os/signal", "Notify").Call(Id("sigChan"), Qual("syscall", "SIGTERM")),
		Qual("os/signal", "Notify").Call(Id("sigChan"), Qual("syscall", "SIGINT")),

		Go().Func().Params().BlockFunc(func(g *Group) {
			g.Op("<-").Id("sigChan")
			if info.IsStopService {
				//TODO add time limit for StopService execution
				g.Id("err").Op(":=").Qual(info.Service.Name, "StopService").Call()
				g.If(Err().Op("!=").Nil()).Block(
					Qual("log", "Println").Call(List(Lit("ERR failed to stop service"), Err())),
				)
			}

			g.For().List(Id("_"), Id("service")).Op(":=").Range().Id("stoppableServices").Block(
				Id("service").Dot("Stop").Call(),
			)

			g.Qual("os", "Exit").Call(Lit(0))
		}).Call(),
	)
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
