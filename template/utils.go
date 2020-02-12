package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

func getMethodTypes(fn *parser.Function, postfix string) (handler, request, response string) {
	method, receiver := fn.Name, fn.Receiver.Type
	handler = fmt.Sprintf("%s%s%sHandler", receiver, method, postfix)
	request = fmt.Sprintf("%s%s%sRequest", receiver, method, postfix)
	response = fmt.Sprintf("%s%s%sResponse", receiver, method, postfix)
	return
}

func isArgNameAreDTO(name string) bool {
	n := strings.ToLower(name)
	return n == "requestdto" || n == "responsedto"
}

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	return strings.ToLower(
		matchAllCap.ReplaceAllString(str, "${1}_${2}"),
	)
}

func getReceiverVarName(receiverTypeName string) string {
	if receiverTypeName == "" {
		return ""
	}
	return fmt.Sprintf("Receiver%s", receiverTypeName)
}

func hasReceiver(fn *parser.Function) bool {
	return fn.Receiver.Type != ""
}

//hasTopLevelReceiver returns true if construcotor has other receiver as argumenet.
func hasTopLevelReceiver(fn *parser.Function, info *PackageInfo) bool {
	if fn == nil {
		return false
	}
	for _, field := range fn.Arguments {
		if _, ok := info.Constructors[field.Type]; ok {
			return false
		}
	}
	return true
}

func forEachFunction(info *PackageInfo, skipInit bool, cb func(*parser.Function)) {
	fns := info.Functions
	if skipInit {
		fns = getFnsWithoutConstructors(info)
	}
	for _, fn := range fns {
		if fn.Name == "Stop" {
			continue
		}
		cb(fn)
	}

}

//getFnsWithoutConstructors removes type constructors
func getFnsWithoutConstructors(info *PackageInfo) (filtered []*parser.Function) {
	fns := info.Functions

	//Get all constructors
	constructors := make(map[*parser.Function]bool)
	for _, fn := range info.Constructors {
		constructors[fn] = true
	}

	for _, fn := range fns {
		if !constructors[fn] {
			filtered = append(filtered, fn)
		}
	}
	return
}

var getTypeFromConstructorName = regexp.MustCompile(`\ANew(.*)\z`)

func isConventionalConstructor(fn *parser.Function) (ok bool, _type string) {
	if hasReceiver(fn) {
		return
	}

	rets := make(map[string]bool)
	for _, ret := range fn.Results {
		rets[ret.Type] = true
	}
	match := getTypeFromConstructorName.FindStringSubmatch(fn.Name)
	if len(match) < 2 {
		return
	}

	return rets[match[1]], match[1]
}

//createCombinedHandlerArgs creates handler arguments that consists of
//original method and constructor arguments (without helpers arguments).
func createCombinedHandlerArgs(fn *parser.Function, info *PackageInfo) []parser.Field {
	arguments := fn.Arguments
	cons := info.GetConstructor(fn.Receiver.Type)
	if cons != nil && !hasTopLevelReceiver(cons, info) {
		for _, arg := range cons.Arguments {
			//Don't include heplers
			if info.IsReceiverType(arg.Type) || arg.Name == "getHeader" || arg.Name == "getEnv" {
				continue
			}
			arguments = append(arguments, arg)
		}
	}
	return arguments
}

func trimPrefix(str string) string {
	return strings.TrimPrefix(str, "*")
}

var matchFuncType = regexp.MustCompile("^func.*")

func getConstructorDepsSignature(fn *parser.Function, info *PackageInfo) (code Code) {
	return getConstructorDeps(fn, info, func(field parser.Field, g *Group) {
		g.Id(getReceiverVarName(field.Type)).Op("*").Qual(info.Service.Name, trimPrefix(field.Type))
	})
}

func getConstructorDeps(
	fn *parser.Function,
	info *PackageInfo,
	createDep func(field parser.Field, g *Group),
) (code Code) {
	if fn == nil {
		return
	}

	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			t := field.Type
			if matchFuncType.MatchString(t) || !info.IsReceiverType(t) {
				continue
			}
			createDep(field, g)
		}
	})
}

func createArgsListFunc(args []parser.Field, params ...string) func(*Group) {
	return createArgsList(args, func(arg *Statement, field parser.Field) *Statement {
		return arg
	}, params...)
}

//createArgsList creates list from parser.Field array.
//Transform function are used to modify each element list.
//Optional param 1 is used to specify prefix for each element.
//Optional param 2 is used to specify allowed argument types (format: type1,type2,).
func createArgsList(
	args []parser.Field,
	transform func(*Statement, parser.Field) *Statement,
	params ...string,
) func(*Group) {
	prefix, onlyTypes := "", ""
	if len(params) > 0 {
		prefix = params[0]
	}
	if len(params) > 1 {
		onlyTypes = params[1]
	}
	return func(g *Group) {
		for _, arg := range args {
			//Skip iteration if arg has type that not in onlyTypes (if it is not empty).
			if onlyTypes != "" && !strings.Contains(onlyTypes, arg.Type+",") {
				continue
			}
			if isArgNameAreDTO(arg.Name) && prefix != "" {
				g.Add(transform(Id(prefix).Dot(arg.Type), arg))
				return
			}
			name := strings.Title(arg.Name)
			if prefix != "" {
				g.Add(transform(Id(prefix).Dot(name), arg))
			} else {
				g.Add(transform(Id(name), arg))
			}
		}
	}
}

func createReqRespTypes(postfix string, info *PackageInfo) Code {
	code := Comment(fmt.Sprintf("Request/Response types (%s)", postfix)).Line()

	forEachFunction(info, true, func(fn *parser.Function) {
		arguments := createCombinedHandlerArgs(fn, info)

		_, reqName, respName := getMethodTypes(fn, postfix)
		code.Add(createTypeFromArgs(reqName, arguments, info))
		code.Line()
		code.Add(createTypeFromArgs(respName, fn.Results, info))
		code.Line()
	})
	return code
}

func createTypeFromArgs(name string, args []parser.Field, info *PackageInfo) Code {
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, arg := range args {
			name := arg.Name
			if isArgNameAreDTO(name) {
				name = ""
			}
			field := Id(strings.Title(name)).Op(arg.Prefix)
			if arg.Package != "" {
				field.Qual(info.Service.Name, arg.Type)
			} else {
				field.Id(arg.Type)
			}
			jsonTag := strings.ToLower(name)
			if arg.Type == "error" {
				jsonTag = "-"
			}
			field.Tag(map[string]string{"json": jsonTag})
			g.Add(field)
		}
	})
}

func injectOriginalMethodCall(g *Group, fn *parser.Function, method Code) {
	g.ListFunc(createArgsListFunc(fn.Results, "response")).
		Op("=").Add(method).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
}

func makeReceiverInitialization(recId string, scope *Group, constructor *parser.Function, info *PackageInfo) {
	if constructor == nil {
		return
	}

	constructorCall := makeCallWithMiddleware(constructor, info, middlewaresMap{"getEnv": Id(getEnvHelper)})

	scope.If(
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, constructor.Name).CallFunc(constructorCall),
		Err().Op("!=").Nil(),
	).Block(
		//TODO return appropriate error here
		Panic(Err()),
	)

	for _, fn := range info.Functions {
		if fn.Name == "Stop" && info.GetConstructor(fn.Receiver.Type) == constructor {
			scope.Id("stoppableServices").Op("=").Append(Id("stoppableServices"), Id(recId))
			return
		}
	}

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

func makeWaitGuard(main *Group) {
	main.Op("<-").Make(Chan().Bool())
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

const getEnvHelper = "getEnvHelper"

func addGetEnvHelper(f *File) {
	f.Func().Id(getEnvHelper).Params(Id("envName").String()).String().Block(
		Return(Qual("os", "Getenv").Call(Id("envName"))),
	)
}

func addIfErrorGuard(scope *Group, statement *Statement, code Code) {
	scope.If(
		statement,
		Err().Op("!=").Nil(),
	).Block(
		Return(code),
	)
}

type middlewaresMap = map[string]*Statement

func makeCallWithMiddleware(fn *parser.Function, info *PackageInfo, middlewares middlewaresMap) func(g *Group) {
	return createArgsList(fn.Arguments, func(arg *Statement, field parser.Field) *Statement {
		fieldName := field.Name

		for name, middleware := range middlewares {
			if fieldName == name {
				return middleware
			}
		}

		if info.IsReceiverType(field.Type) {
			return Id(getReceiverVarName(field.Type))
		}

		//Oterwise inject receiver dependencie
		//TODO send nil for pointer or empty object otherwise
		//TODO why listfunc instead of g.Id?
		return ListFunc(createArgsListFunc([]parser.Field{field}, "request"))
	})
}
