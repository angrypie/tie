package template

import (
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
	. "github.com/dave/jennifer/jen"
)

//getConstructorDepsSignature creates constructor signature list.
func getConstructorDepsSignature(constructor Constructor, info *PackageInfo) (code Code) {
	return getConstructorDeps(constructor, info, func(field parser.Field, g *Group) {
		typeName := field.TypeName()
		g.Id(GetReceiverVarName(typeName)).Add(createTypeFromField(field, info))
	})
}

//getConstructorDeps creates list that contain constructor args, except function
//and receiver types. createDep is used to construct single list element.
func getConstructorDeps(
	constructor Constructor,
	info *PackageInfo,
	createDep func(field parser.Field, g *Group),
) (code Code) {
	fn := constructor.Function

	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			t := field.TypeName()
			if isFuncType(t) || !info.IsReceiverType(field) {
				continue
			}
			createDep(field, g)
		}
	})
}

//CreateArgsListFunc creates list from Fields (see CreateArgListFunc).
func CreateArgsListFunc(args []parser.Field, params ...string) func(*Group) {
	return CreateArgsList(args, func(arg *Statement, field parser.Field) *Statement {
		return arg
	}, params...)
}

//CreateSignatureFromArgs creates signature from Fields (see CreateArgListFunc).
func CreateSignatureFromArgs(args []parser.Field, info *PackageInfo, params ...string) func(*Group) {
	return CreateArgsList(args, func(arg *Statement, field parser.Field) *Statement {
		return Id(field.Name()).Add(createTypeFromField(field, info))
	}, params...)
}

//CreateArgsList creates list from parser.Field array.
//Transform function is used to modify each element list.
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
			if onlyTypes != "" && !strings.Contains(onlyTypes, arg.TypeName()+",") {
				continue
			}
			if isArgNameAreDTO(arg.Name()) && prefix != "" {
				g.Add(transform(Id(prefix).Dot(arg.TypeName()), arg))
				return
			}

			if prefix != "" {
				name := strings.Title(arg.Name())
				g.Add(transform(Id(prefix).Dot(name), arg))
			} else {
				g.Add(transform(Id(arg.Name()), arg))
			}
		}
	}
}

//CreateTypeAliases creates aliases to types that found in functions signatures.
func CreateTypeAliases(info *PackageInfo, f *File) {
	f.Comment("Type aliases")
	done := make(map[string]bool)
	ForEachFunction(info, true, func(fn parser.Function) {
		fields := append(fn.Arguments, fn.Results.List()...)
		for _, field := range fields {
			//Skip not local types and already processed types
			_, path, local := field.TypeParts()
			if info.Service.Name != path || done[local] {
				continue
			}
			done[local] = true
			f.Type().Id(local).Op("=").Qual(info.GetServicePath(), local)
			f.Line()
		}
	})
}

//CreateReqRespTypes creates request response types for each method.
func CreateReqRespTypes(info *PackageInfo, f *File) {
	f.Comment("Request/Response types")
	cb := func(receiver parser.Field, constructor OptionalConstructor) {
		t, c := ClientReceiverType(receiver, constructor, info)
		f.Add(t).Line().Add(c).Line()
	}
	MakeForEachReceiver(info, cb)

	ForEachFunction(info, true, func(fn parser.Function) {
		arguments := CreateCombinedHandlerArgs(fn, info)
		results := fieldsFromParser(fn.Results.List())

		_, reqName, respName := GetMethodTypes(fn)
		f.Add(TypeDeclFormFields(reqName, arguments, info))
		f.Line()
		f.Add(TypeDeclFormFields(respName, results, info))
		f.Line()
	})
}

//TODO replace to to Struct type
//TypeDeclFormFields creates type declaration from []types.Field
func TypeDeclFormFields(name string, args []types.Field, info *PackageInfo) Code {
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, arg := range args {
			name := arg.Name()
			if isArgNameAreDTO(name) {
				name = ""
			}
			field := Id(strings.Title(name)).Add(createTypeFromField(arg, info))
			jsonTag := strings.ToLower(name)
			if arg.TypeName() == "error" {
				jsonTag = "-"
			}
			field.Tag(map[string]string{"json": jsonTag})
			g.Add(field)
		}
	})
}

//ClientReceiverType creates constructor and client-side receiver type.
//Type contains only fields from constructor arguments. Contructor match
//original one by signature but only initializes recevier fieds.
//Example: type Foo{...}; NewFoo(x int) -> type Foo { x int }; NewFoo(x int)
func ClientReceiverType(receiver parser.Field, constructor OptionalConstructor, info *PackageInfo) (
	typeDecl, constructorDecl Code) {
	receiverType := receiver.TypeName()

	constructor(func(c Constructor) {
		fn := c.Function
		args, results := fn.Arguments, fn.Results.List()

		typeDecl = Type().Id(receiverType).StructFunc(func(g *Group) {
			for _, arg := range filterHelperArgs(args, info) {
				field := Id(strings.Title(arg.Name())).Add(createTypeFromField(arg, info))
				g.Add(field)
			}
		})

		transformSignature := func(fields []parser.Field) func(*Group) {
			return CreateArgsList(fields, func(arg *Statement, field parser.Field) *Statement {
				if _, ok := info.GetConstructor(field); ok {
					prefix, _, local := field.TypeParts()
					return Id(field.Name()).Id(prefix + local)
				}
				return Id(field.Name()).Add(createTypeFromField(field, info))
			})
		}

		constructorDecl = Func().Id(fn.Name).
			ParamsFunc(transformSignature(args)).
			ParamsFunc(transformSignature(results)).
			BlockFunc(func(g *Group) {
				//TODO do not gues but find returned receiver by type
				receiver := results[0].Name()
				g.Id(receiver).Op("=").New(Id(receiverType))

				filtered := filterHelperArgs(args, info)
				g.ListFunc(CreateArgsListFunc(filtered, receiver)).Op("=").
					ListFunc(CreateArgsListFunc(filtered))

				g.Return(ListFunc(CreateArgsListFunc(results)))
			})
	}, func() {
		typeDecl = Type().Id(receiverType).Struct()
	})
	return
}

//createTypeFromField create qualified type from types.Field.
func createTypeFromField(field types.Field, info *PackageInfo) Code {
	prefix, path, local := field.TypeParts()
	if path == "" {
		return Op(local)
	}
	if path == info.Service.Name {
		path = info.GetServicePath()
	}
	return Op(prefix).Qual(path, local)
}

//injectOriginalMethodCall injects original method call.
func injectOriginalMethodCall(g *Group, fn parser.Function, method Code) {
	g.ListFunc(CreateArgsListFunc(fn.Results.List(), "response")).
		Op("=").Add(method).Call(ListFunc(CreateArgsListFunc(fn.Arguments, "request")))
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

//TODO
//makeWaitGuard creates guard that blocks current thread
func makeWaitGuard(main *Group) {
	main.Op("<-").Make(Chan().Bool())
}

//GracefulShutdown calls Stop methods on receivers before program termination.
func GracefulShutdown(info *PackageInfo, g *Group, f *File) {
	g.Comment("GracefulShutdown(local scope)").Line()
	f.Comment("GracefulShutdown (file)").Line()

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
	return
}

//GetEnvHelper global identifier for getEnv helper function.
const GetEnvHelper = "getEnvHelper"

//AddGetEnvHelper creates heper function to get environment variable.
func AddGetEnvHelper(f *File) {
	f.Func().Id(GetEnvHelper).Params(Id("envName").String()).String().Block(
		Return(Qual("os", "Getenv").Call(Id("envName"))),
	)
}

type IfErrorGuard = func(scope *Group, statement *Statement)

//AddIfErrorGuard adds to scope error check.
func AddIfErrorGuard(scope *Group, statement *Statement, errId string, code Code) {
	scope.If(
		statement,
		Id(errId).Op("!=").Nil(),
	).Block(
		Return(code),
	)
}

//BindErrToResults assigns err statement to last item in fields if it's error.
func AssignErrToResults(err *Statement, fields parser.ResultFields) (statement *Statement) {
	last := fields.Last
	if last.TypeName() == "error" {
		return Id(last.Name()).Op("=").Add(err)
	}
	return
}

//AssignResultsToErr assign response error to err statement.
func AssignResultsToErr(err *Statement, respId string, fields parser.ResultFields) (statement *Statement) {
	last := fields.Last
	if last.TypeName() != "error" {
		return
	}
	return err.Op("=").ListFunc(CreateArgsListFunc([]parser.Field{last}, respId))
}

type MiddlewaresMap = map[string]*Statement

//makeCallWithMiddleware injects middlewares to args list for constructor.
func makeCallWithMiddleware(constructor Constructor, info *PackageInfo, middlewares MiddlewaresMap) func(g *Group) {
	return CreateArgsList(constructor.Function.Arguments, func(arg *Statement, field parser.Field) *Statement {
		fieldName := field.Name()

		for name, middleware := range middlewares {
			if fieldName == name {
				return middleware
			}
		}

		//Inject receiver dependencie
		if info.IsReceiverType(field) {
			return Id(GetReceiverVarName(field.TypeName()))
		}

		if isFuncType(field.TypeName()) {
			return Nil()
		}

		//TODO send nil for pointer or empty object
		//Bind request argument
		return ListFunc(CreateArgsListFunc([]parser.Field{field}, "request."+RequestReceiverKey))
	})
}

//makeEmptyValuesMiddlewareCall inject middlewares and empy values to args list for constructor.
func makeEmptyValuesMiddlewareCall(fn parser.Function, info *PackageInfo, middlewares MiddlewaresMap) func(g *Group) {
	return CreateArgsList(fn.Arguments, func(arg *Statement, field parser.Field) *Statement {
		fieldName := field.Name()
		//TODO CHECK
		prefix, path, local := field.TypeParts()

		for name, middleware := range middlewares {
			if fieldName == name {
				return middleware
			}
		}

		//TODO deduct zero value from any type
		switch field.TypeName() {
		case "string":
			return Lit("")
		case "int":
			return Lit(0)
		}

		//TODO add isInterface check remove IsExported from condition and uncomment lines bellow
		if isFuncType(local) || prefix != "" || path != "" {
			return Nil()
		}

		//if ast.IsExported(fieldType) {
		//return Qual(info.GetServicePath(), TrimPrefix(fieldType)).Block()
		//}

		return Id(local)
	})
}

const rndport = "github.com/angrypie/rndport"

//MakeStartServerInit creates port and address initialization (from env or random).
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
			AddIfErrorGuard(g, nil, "err", nil)
		} else {
			g.Id("portStr").Op("=").Lit(portStr)
		}
	})
	g.List(Id("port"), Err()).Op(":=").Qual("strconv", "Atoi").Call(Id("portStr"))
	g.Id("_").Op("=").Id("port")
	AddIfErrorGuard(g, nil, "err", nil)
	g.Id("address").Op(":=").Lit("0.0.0.0:").Op("+").Id("portStr")
}

//MakeReceiversForHandlers cerates instances for each top level receiver.
func MakeReceiversForHandlers(info *PackageInfo, g *Group) (receiversCreated map[string]parser.Field) {
	receiversCreated = make(map[string]parser.Field)
	cb := func(receiver parser.Field, constructor OptionalConstructor) {

		receiverType := receiver.TypeName()
		recId := GetReceiverVarName(receiverType)
		var skipInitStopable bool

		//creates receiver instance using constructor if it exist, othewise using new().
		constructor(
			func(c Constructor) {
				//Skip not top level receivers.
				if !HasTopLevelReceiver(c.Function, info) {
					skipInitStopable = true
					return
				}
				fn := c.Function
				constructorCall := makeEmptyValuesMiddlewareCall(fn, info, MiddlewaresMap{"getEnv": Id(GetEnvHelper)})
				g.List(Id(recId), Err()).Op(":=").Qual(info.GetServicePath(), fn.Name).CallFunc(constructorCall)
				AddIfErrorGuard(g, nil, "err", nil)

				receiversCreated[receiver.TypeName()] = receiver
			}, func() {
				g.Id(recId).Op(":=").New(Qual(info.GetServicePath(), receiverType))
				receiversCreated[receiver.TypeName()] = receiver
			})

		if skipInitStopable {
			return
		}
		for _, fn := range info.Functions {
			if fn.Name == "Stop" {
				g.Id("stoppableServices").Op("=").Append(Id("stoppableServices"), Id(recId))
				return
			}
		}
	}
	MakeForEachReceiver(info, cb)
	return receiversCreated
}

//MakeHandlerWrapperCall creates args list for HandlerWrapper.
func MakeHandlerWrapperCall(fn parser.Function, info *PackageInfo, createDep func(string) Code) func(*Group) {
	return func(g *Group) {
		if !HasReceiver(fn) {
			return
		}
		constructor, ok := info.GetConstructor(fn.Receiver)

		if !ok || HasTopLevelReceiver(constructor.Function, info) {
			//Inject receiver to http handler.
			receiverVarName := GetReceiverVarName(fn.Receiver.TypeName())
			g.Add(createDep(receiverVarName))
		} else {
			//Inject dependencies to handler for non top level receiver.
			g.Add(getConstructorDeps(constructor, info, func(field parser.Field, g *Group) {
				receiverVarName := GetReceiverVarName(field.TypeName())
				g.Add(createDep(receiverVarName))
			}))
		}
	}
}

//TODO accept Statement insteal Group
//MakeOriginalCall creates dependencies and make original method call (response object must be created)
func MakeOriginalCall(
	info *PackageInfo, fn parser.Function, g *Group,
	middlewares MiddlewaresMap, errGuard IfErrorGuard,
) {
	//If method has receiver generate receiver middleware code
	//else just call public package method
	if HasReceiver(fn) {
		constructor, ok := info.GetConstructor(fn.Receiver)
		receiverType := fn.Receiver.TypeName()
		//TODO replace recId with generated name
		recId := GetReceiverVarName(receiverType)
		if ok && !HasTopLevelReceiver(constructor.Function, info) {
			g.Id(recId).Op(":=").New(Qual(info.GetServicePath(), receiverType))

			constructorCall := makeCallWithMiddleware(constructor, info, middlewares)
			errGuard(g, List(Id(recId), Err()).Op("=").
				Qual(info.GetServicePath(), constructor.Function.Name).CallFunc(constructorCall),
			)
		}
		injectOriginalMethodCall(g, fn, Id(recId).Dot(fn.Name))
	} else {
		injectOriginalMethodCall(g, fn, Qual(info.GetServicePath(), fn.Name))
	}
	errGuard(g, AssignResultsToErr(Err(), "response", fn.Results))
}

//HandlerWrapper creates method wrapper to inject dependencies (top level receiver).
func MakeHandlerWrapper(
	f *File, handlerBody func(g *Group), info *PackageInfo, fn parser.Function,
	args, returns *Statement,
) {
	handler, _, _ := GetMethodTypes(fn)

	wrapperParams := func(g *Group) {
		if !HasReceiver(fn) {
			return
		}
		receiverVarName := GetReceiverVarName(fn.Receiver.TypeName())
		constructor, ok := info.GetConstructor(fn.Receiver)
		if !ok || HasTopLevelReceiver(constructor.Function, info) {
			_, _, recLocal := fn.Receiver.TypeParts()
			g.Id(receiverVarName).Op("*").Qual(info.GetServicePath(), recLocal)
		} else {
			g.Add(getConstructorDepsSignature(constructor, info))
		}
	}

	f.Func().Id(handler).ParamsFunc(wrapperParams).Func().Params(args).Params(returns).Block(
		Return(Func().Params(args).Params(returns).BlockFunc(func(g *Group) {
			handlerBody(g)
		})),
	).Line()
}
