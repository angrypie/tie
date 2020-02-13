package template

import (
	"fmt"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

const rpcxServer = "github.com/smallnest/rpcx/server"
const rndport = "github.com/angrypie/rndport"

func GetServerMainRPC(info *PackageInfo) (string, error) {
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
	main.Go().Id("startServer").Call()

	f.Func().Id("startServer").Params().BlockFunc(func(g *Group) {
		makeStartServerInit(info, g) //SIM
		receiversCreated := makeReceiversForHandlers(info, g)

		//RC replace http server init
		resourceName := getRPCResourceName(info)
		resourceInstance := "Instance__" + resourceName

		f.Type().Id(resourceName).StructFunc(func(g *Group) {
			for receiverType := range receiversCreated {
				receiverVarName := getReceiverVarName(receiverType)
				g.Id(receiverVarName).Op("*").Qual(info.Service.Name, trimPrefix(receiverType))
			}
		})

		//.2 Add handler for each function.
		forEachFunction(info, true, func(fn *parser.Function) {
			handler, request, response := getMethodTypes(fn, "RPC")

			f.Func().Params(Id("resource").Id(resourceName)).Id(handler).
				Params(Id("ctx").Qual("context", "Context"), Id("request").Id(request), Id("response").Id(response)).
				Params(Err().Error()).Block(
				Return(
					Id(handler).
						CallFunc(makeHandlerWrapperCall(fn, info)).
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
