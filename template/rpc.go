package template

import (
	"fmt"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

func GetRpcServerMain(info *PackageInfo) (string, error) {
	f := NewFile("rpc")

	f.Func().Id("Main").Params().BlockFunc(func(g *Group) {
		makeGracefulShutdown(info, g, f)
		makeInitService(info, g, f)

		makeRPCServer(info, g, f)
	})

	return fmt.Sprintf("%#v", f), nil
}

func makeRPCServer(info *PackageInfo, main *Group, f *File) {
	makeStartRPCServer(info, main, f)
	makeRPCRequestResponseTypes(info, main, f)
	makeRPCHandlers(info, main, f)
	makeHelpersRPC(f)
}

func makeStartRPCServer(info *PackageInfo, main *Group, f *File) {
	rpcxServer := "github.com/smallnest/rpcx/server"
	rndport := "github.com/angrypie/rndport"

	main.Go().Id("startRPCServer").Call()

	f.Func().Id("startRPCServer").Params().BlockFunc(func(g *Group) {
		//Declare err and get rid of ''unused' error.
		g.Var().Err().Error()
		g.Id("_").Op("=").Err()

		port := info.Service.Port
		if port == "" {
			g.List(Id("address"), Err()).Op(":=").
				Qual(rndport, "GetAddress").Call(Lit(":%d"))
			g.If(Err().Op("!=").Nil()).Block(Panic(Err()))
		} else {
			g.Id("address").Op(":=").Lit(fmt.Sprintf(":%s", port))
		}

		//. Set HTTP handlers and init receivers.
		//.1 Create receivers for handlers
		receiversProcessed := make(map[string]bool)
		receiversCreated := make(map[string]bool) //RC
		createReceivers := func(receiverType string, constructorFunc *parser.Function) {
			receiversProcessed[receiverType] = true
			//Skip not top level receivers.
			if constructorFunc != nil && !hasTopLevelReceiver(constructorFunc, info) {
				return
			}
			receiversCreated[receiverType] = true
			receiverVarName := getReceiverVarName(receiverType)
			g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, trimPrefix(receiverType)).Block()
			makeReceiverInitialization(receiverVarName, g, constructorFunc, info)
		}
		//Create receivers for each constructor
		for t, c := range info.Constructors {
			createReceivers(t, c)
		}

		//Create receivers that does not have constructor
		forEachFunction(info, false, func(fn *parser.Function) {
			receiverType := fn.Receiver.Type
			//Skip function if it does not have receiver or receiver already created.
			if !hasReceiver(fn) || receiversProcessed[receiverType] {
				return
			}
			//It will not create constructor call due constructor func is nil
			createReceivers(receiverType, nil)
		})

		//RC replace http server init
		resourceName := getRPCResourceName(info)
		resourceInstance := "Instance__" + resourceName

		f.Type().Id(resourceName).StructFunc(func(g *Group) {
			for receiverType := range receiversCreated {
				receiverVarName := getReceiverVarName(receiverType)
				g.Id(receiverVarName).Op("*").Qual(info.Service.Name, trimPrefix(receiverType))
			}
		})

		//RC add rpc handlers for main resource object
		forEachFunction(info, true, func(fn *parser.Function) {
			handler, request, response := getMethodTypes(fn, "RPC")
			constructorFunc := info.GetConstructor(fn.Receiver.Type)
			receiverVarName := getReceiverVarName(fn.Receiver.Type)

			f.Func().Params(Id("resource").Id(resourceName)).Id(handler).
				Params(Id("ctx").Qual("context", "Context"), Id("request").Id(request), Id("response").Id(response)). //RC
				Params(Err().Error()).Block(
				Return(Id(handler).
					CallFunc(func(g *Group) {
						if constructorFunc == nil {
							return
						}
						if hasTopLevelReceiver(constructorFunc, info) {
							//Inject receiver to http handler.
							g.Id("resource").Dot(receiverVarName)
						} else {
							//Inject dependencies to rpc handler for non top level receiver.
							g.Add(getConstructorDeps(constructorFunc, info, func(field parser.Field, g *Group) {
								g.Id("resource").Dot(getReceiverVarName(field.Type))
							}))
						}
					}).
					Call(Id("ctx"), Id("request"), Id("response"))))
		})

		g.Id("server").Op(":=").Qual(rpcxServer, "NewServer").Call()
		g.Id(resourceInstance).Op(":=").Op("&").Id(resourceName).Values(DictFunc(func(d Dict) {
			for receiverType := range receiversCreated {
				receiverVarName := getReceiverVarName(receiverType)
				d[Id(receiverVarName)] = Id(receiverVarName)
			}
		}))
		g.Id("server").Dot("RegisterName").Call(Lit(resourceName), Id(resourceInstance), Lit(""))

		g.Id("server").Dot("Serve").Call(Lit("tcp"), Id("address"))

		//RC end
	})
}

func makeRPCRequestResponseTypes(info *PackageInfo, main *Group, f *File) {
	f.Add(createReqRespTypes("RPC", info))
}

func makeRPCHandlers(info *PackageInfo, main *Group, f *File) {
	f.Comment(fmt.Sprintf("API handler methods (%s)", "RPC")).Line()
	forEachFunction(info, true, func(fn *parser.Function) {
		makeRPCHandler(info, fn, f.Group)
	})
}

func makeRPCHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	handler, request, response := getMethodTypes(fn, "RPC")
	receiverVarName := getReceiverVarName(fn.Receiver.Type)
	handlerBody := func(g *Group) {
		//Bind request params RC deleted request  binding

		//Create response object RC deleted

		//If method has receiver generate receiver middleware code
		//else just call public package method
		if hasReceiver(fn) {
			constructorFunc := info.GetConstructor(fn.Receiver.Type)
			if constructorFunc != nil && !hasTopLevelReceiver(constructorFunc, info) {
				receiverType := fn.Receiver.Type
				g.Id(receiverVarName).Op(":=").Op("&").Qual(info.Service.Name, trimPrefix(receiverType)).Block()
				makeReceiverMiddlewareRPC(receiverVarName, g, constructorFunc, info) //RC
			}
			injectOriginalMethodCall(g, fn, Id(receiverVarName).Dot(fn.Name))
		} else {
			injectOriginalMethodCall(g, fn, Qual(info.Service.Name, fn.Name))
		}

		ifErrorReturnErrRPC( //RC return rpc compatible error from handler
			g,
			Err().Op(":=").Id("response").Dot("Err"),
		)

		g.Return(Nil())
	}

	//Create handler methods that use closure to inject receiver if it exist.
	file.Func().Id(handler).ParamsFunc(func(g *Group) {
		if !hasReceiver(fn) {
			return
		}
		constructorFunc := info.GetConstructor(fn.Receiver.Type)
		if constructorFunc == nil || hasTopLevelReceiver(constructorFunc, info) {
			g.Id(receiverVarName).Op("*").Qual(info.Service.Name, trimPrefix(fn.Receiver.Type))
		} else {
			g.Add(getConstructorDepsSignature(constructorFunc, info))
		}
	}).Params(
		Func().Params(Qual("context", "Context"), Id(request), Id(response)).Params(Error()), //RC
	).Block(Return(Func().
		Params(Id("ctx").Qual("context", "Context"), Id("request").Id(request), Id("response").Id(response)). //RC
		Params(Err().Error()).BlockFunc(handlerBody),
	)).Line()
}

func ifErrorReturnErrRPC(scope *Group, statement *Statement) {
	addIfErrorGuard(scope, statement, Err())
}

func makeReceiverMiddlewareRPC(recId string, scope *Group, constructor *parser.Function, info *PackageInfo) {
	if constructor == nil {
		return
	}
	constructorCall := makeCallWithMiddleware(constructor, info, middlewaresMap{"getEnv": Id(getEnvHelper)})

	ifErrorReturnErrRPC(
		scope,
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, constructor.Name).CallFunc(constructorCall),
	)
}

func getRPCResourceName(info *PackageInfo) string {
	return "Resource__" + info.Service.Alias
}

func makeHelpersRPC(f *File) {
	addGetEnvHelper(f)
}
