package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

func GetMethodTypes(fn *parser.Function, postfix string) (handler, request, response string) {
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

func GetReceiverVarName(receiverTypeName string) string {
	if receiverTypeName == "" {
		return ""
	}
	return fmt.Sprintf("Receiver%s", receiverTypeName)
}

func HasReceiver(fn *parser.Function) bool {
	return fn.Receiver.Type != ""
}

//HasTopLevelReceiver returns true if constructor has other receiver as argumenet.
func HasTopLevelReceiver(fn *parser.Function, info *PackageInfo) bool {
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

func ForEachFunction(info *PackageInfo, skipInit bool, cb func(*parser.Function)) {
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
	if HasReceiver(fn) {
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

func TrimPrefix(str string) string {
	return strings.TrimPrefix(str, "*")
}

var matchFuncType = regexp.MustCompile("^func.*")

func isFuncType(t string) bool {
	return matchFuncType.MatchString(t)
}

func getConstructorDepsSignature(fn *parser.Function, info *PackageInfo) (code Code) {
	return getConstructorDeps(fn, info, func(field parser.Field, g *Group) {
		g.Id(GetReceiverVarName(field.Type)).Op("*").Qual(info.GetServicePath(), TrimPrefix(field.Type))
	})
}

func getConstructorDeps(
	fn *parser.Function,
	info *PackageInfo,
	createDep func(field parser.Field, g *Group),
) (code Code) {
	if fn == nil {
		return List()
	}

	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			t := field.Type
			if isFuncType(t) || !info.IsReceiverType(t) {
				continue
			}
			createDep(field, g)
		}
	})
}

func CreateArgsListFunc(args []parser.Field, params ...string) func(*Group) {
	return CreateArgsList(args, func(arg *Statement, field parser.Field) *Statement {
		return arg
	}, params...)
}

func CreateSignatureFromArgs(args []parser.Field, params ...string) func(*Group) {
	return CreateArgsList(args, func(arg *Statement, field parser.Field) *Statement {
		return Id(field.Name).Id(field.Type)
	}, params...)
}

//CreateArgsList creates list from parser.Field array.
//Transform function are used to modify each element list.
//Optional param 1 is used to specify prefix for each element.
//Optional param 2 is used to specify allowed argument types (format: type1,type2,).
func CreateArgsList(
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
			if prefix != "" {
				name := strings.Title(arg.Name)
				g.Add(transform(Id(prefix).Dot(name), arg))
			} else {
				g.Add(transform(Id(arg.Name), arg))
			}
		}
	}
}

func CreateReqRespTypes(postfix string, info *PackageInfo) Code {
	code := Comment(fmt.Sprintf("Request/Response types (%s)", postfix)).Line()

	ForEachFunction(info, true, func(fn *parser.Function) {
		arguments := CreateCombinedHandlerArgs(fn, info)

		_, reqName, respName := GetMethodTypes(fn, postfix)
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
				field.Qual(info.GetServicePath(), arg.Type)
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
	g.ListFunc(CreateArgsListFunc(fn.Results, "response")).
		Op("=").Add(method).Call(ListFunc(CreateArgsListFunc(fn.Arguments, "request")))
}

func makeReceiverInitialization(receiverType string, g *Group, constructor *parser.Function, info *PackageInfo) {
	recId := GetReceiverVarName(receiverType)
	if constructor == nil {
		g.Id(recId).Op(":=").Op("&").Qual(info.GetServicePath(), TrimPrefix(receiverType)).Block()
		return
	}

	constructorCall := makeCallWithMiddleware(constructor, info, MiddlewaresMap{"getEnv": Id(GetEnvHelper)})
	g.List(Id(recId), Err()).Op(":=").Qual(info.GetServicePath(), constructor.Name).CallFunc(constructorCall)
	AddIfErrorGuard(g, nil, nil)

	for _, fn := range info.Functions {
		if fn.Name == "Stop" && info.GetConstructor(fn.Receiver.Type) == constructor {
			g.Id("stoppableServices").Op("=").Append(Id("stoppableServices"), Id(recId))
			return
		}
	}
}

func MakeInitService(info *PackageInfo, main *Group) {
	if !info.IsInitService {
		return
	}
	main.If(
		Err().Op(":=").Qual(info.GetServicePath(), "InitService").Call(),
		Err().Op("!=").Nil(),
	).Block(
		createErrLog("failed to init service"),
		Return(),
	)
}

func makeWaitGuard(main *Group) {
	main.Op("<-").Make(Chan().Bool())
}

func MakeGracefulShutdown(info *PackageInfo, g *Group, f *File) {
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
				g.Id("err").Op(":=").Qual(info.GetServicePath(), "StopService").Call()
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

const GetEnvHelper = "getEnvHelper"

func AddGetEnvHelper(f *File) {
	f.Func().Id(GetEnvHelper).Params(Id("envName").String()).String().Block(
		Return(Qual("os", "Getenv").Call(Id("envName"))),
	)
}

type IfErrorGuard = func(scope *Group, statement *Statement)

func AddIfErrorGuard(scope *Group, statement *Statement, code Code) {
	scope.If(
		statement,
		Err().Op("!=").Nil(),
	).Block(
		Return(code),
	)
}

type MiddlewaresMap = map[string]*Statement

func makeCallWithMiddleware(fn *parser.Function, info *PackageInfo, middlewares MiddlewaresMap) func(g *Group) {
	return CreateArgsList(fn.Arguments, func(arg *Statement, field parser.Field) *Statement {
		fieldName := field.Name

		for name, middleware := range middlewares {
			if fieldName == name {
				return middleware
			}
		}

		//Inject receiver dependencie
		if info.IsReceiverType(field.Type) {
			return Id(GetReceiverVarName(field.Type))
		}

		if isFuncType(field.Type) {
			return Nil()
		}

		//TODO send nil for pointer or empty object
		//Bind request argument
		return ListFunc(CreateArgsListFunc([]parser.Field{field}, "request"))
	})
}

const rndport = "github.com/angrypie/rndport"

func MakeStartServerInit(info *PackageInfo, g *Group) {
	portStr := info.Service.Port

	//Try to use port value from environment
	g.Var().Id("portStr").String()
	g.If(
		Id("p").Op(":=").Qual("os", "Getenv").Call(Lit("PORT")),
		Id("p").Op("!=").Lit(""),
	).Block(
		Id("portStr").Op("=").Id("p"),
	).Else().BlockFunc(func(g *Group) {
		//Use random port if configuration and environment is empty
		if portStr == "" {
			g.List(Id("portStr"), Err()).Op("=").Qual(rndport, "GetAddress").Call(Lit("%d"))
			AddIfErrorGuard(g, nil, nil)
		} else {
			g.Id("portStr").Op("=").Lit(portStr)
		}
	})
	g.List(Id("port"), Err()).Op(":=").Qual("strconv", "Atoi").Call(Id("portStr"))
	g.Id("_").Op("=").Id("port")
	AddIfErrorGuard(g, nil, nil)
	g.Id("address").Op(":=").Lit("localhost:").Op("+").Id("portStr")
}

//MakeReceiversForHandlers cerates instances for each top level receiver.
func MakeReceiversForHandlers(info *PackageInfo, g *Group) (receiversCreated map[string]bool) {
	receiversCreated = make(map[string]bool)
	cb := func(receiverType string, constructor *parser.Function) {
		//Skip not top level receivers.
		if constructor != nil && !HasTopLevelReceiver(constructor, info) {
			return
		}
		receiversCreated[receiverType] = true
		makeReceiverInitialization(receiverType, g, constructor, info)
	}
	MakeForEachReceiver(info, cb)
	return receiversCreated
}

//MakeForEachReceiver executes callback for each receiver.
func MakeForEachReceiver(
	info *PackageInfo, cb func(receiverType string, constructor *parser.Function),
) (receiversProcessed map[string]bool) {
	//TODO MakeForEachReceiver code was edited from MakeReceiversForHandlers,
	//it may possibly contain redunant checks, loops.
	receiversProcessed = make(map[string]bool)
	createReceivers := func(receiverType string, constructorFunc *parser.Function) {
		receiversProcessed[receiverType] = true
		cb(receiverType, constructorFunc)
	}
	//Create receivers for each constructor
	for t, c := range info.Constructors {
		createReceivers(t, c)
	}

	//Create receivers that does not have constructor
	ForEachFunction(info, false, func(fn *parser.Function) {
		receiverType := fn.Receiver.Type
		//Skip function if it does not have receiver or receiver already created.
		if !HasReceiver(fn) || receiversProcessed[receiverType] {
			return
		}
		//It will not create constructor call due constructor func is nil
		createReceivers(receiverType, nil)
	})

	return receiversProcessed
}

func MakeHandlerWrapperCall(fn *parser.Function, info *PackageInfo, createDep func(string) Code) func(*Group) {
	constructorFunc := info.GetConstructor(fn.Receiver.Type)
	receiverVarName := GetReceiverVarName(fn.Receiver.Type)
	return func(g *Group) {
		if !HasReceiver(fn) {
			return
		}
		if constructorFunc == nil || HasTopLevelReceiver(constructorFunc, info) {
			//Inject receiver to http handler.
			g.Add(createDep(receiverVarName))
		} else {
			//Inject dependencies to http handler for non top level receiver.
			g.Add(getConstructorDeps(constructorFunc, info, func(field parser.Field, g *Group) {
				g.Add(createDep(GetReceiverVarName(field.Type)))
			}))
		}
	}
}

func MakeHandlers(
	info *PackageInfo, f *File,
	makeHandler func(info *PackageInfo, fn *parser.Function, file *Group),
) {
	f.Comment(fmt.Sprintf("API handler methods")).Line()
	ForEachFunction(info, true, func(fn *parser.Function) {
		makeHandler(info, fn, f.Group)
	})
}

//MakeOriginalCall creates dependencies and make original method call (response object must be created)
func MakeOriginalCall(
	info *PackageInfo, fn *parser.Function, g *Group,
	middlewares MiddlewaresMap, errGuard IfErrorGuard,
) {
	//If method has receiver generate receiver middleware code
	//else just call public package method
	if HasReceiver(fn) {
		receiverType := fn.Receiver.Type
		constructor := info.GetConstructor(receiverType)
		receiverVarName := GetReceiverVarName(receiverType)
		if constructor != nil && !HasTopLevelReceiver(constructor, info) {
			g.Id(receiverVarName).Op(":=").Op("&").Qual(info.GetServicePath(), TrimPrefix(receiverType)).Block()

			constructorCall := makeCallWithMiddleware(constructor, info, middlewares)
			errGuard(g, List(Id(receiverVarName), Err()).Op("=").
				Qual(info.GetServicePath(), constructor.Name).CallFunc(constructorCall),
			)
		}
		injectOriginalMethodCall(g, fn, Id(receiverVarName).Dot(fn.Name))
	} else {
		injectOriginalMethodCall(g, fn, Qual(info.GetServicePath(), fn.Name))
	}

	errGuard(g, Err().Op(":=").Id("response").Dot("Err"))
}

func MakeHandlerWrapper(
	moduleId string, handlerBody func(*Group), info *PackageInfo, fn *parser.Function, file *Group,
	args, returns *Statement,
) {
	handler, _, _ := GetMethodTypes(fn, moduleId)
	receiverVarName := GetReceiverVarName(fn.Receiver.Type)

	wrapperParams := func(g *Group) {
		if !HasReceiver(fn) {
			return
		}
		constructorFunc := info.GetConstructor(fn.Receiver.Type)
		if constructorFunc == nil || HasTopLevelReceiver(constructorFunc, info) {
			g.Id(receiverVarName).Op("*").Qual(info.GetServicePath(), TrimPrefix(fn.Receiver.Type))
		} else {
			g.Add(getConstructorDepsSignature(constructorFunc, info))
		}
	}

	file.Func().Id(handler).ParamsFunc(wrapperParams).Func().Params(args).Params(returns).Block(
		Return(Func().Params(args).Params(returns).BlockFunc(handlerBody)),
	).Line()
}

//CreateCombinedHandlerArgs creates handler arguments that consists of
//original method and constructor arguments (without helpers arguments).
func CreateCombinedHandlerArgs(fn *parser.Function, info *PackageInfo) []parser.Field {
	arguments := fn.Arguments
	cons := info.GetConstructor(fn.Receiver.Type)
	if cons != nil && !HasTopLevelReceiver(cons, info) {
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
