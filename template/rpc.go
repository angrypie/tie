package template

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

const rpcxServer = "github.com/smallnest/rpcx/server"
const rndport = "github.com/angrypie/rndport"
const rpcModuleId = "RPC"

func GetServerMainRPC(info *PackageInfo) (string, error) {
	f := NewFile(strings.ToLower(rpcModuleId))

	f.Func().Id("Main").Params().BlockFunc(func(main *Group) {
		makeGracefulShutdown(info, main, f)
		makeInitService(info, main)

		makeStartRPCServer(info, main, f)
	})
	makeHandlers(info, f, makeRPCHandler)
	f.Add(createReqRespTypes(rpcModuleId, info))
	makeHelpersRPC(f)

	return fmt.Sprintf("%#v", f), nil
}

func makeRPCHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	handler, request, response := getMethodTypes(fn, rpcModuleId)
	receiverVarName := getReceiverVarName(fn.Receiver.Type)

	handlerBody := func(g *Group) {
		middlewares := middlewaresMap{"getEnv": Id(getEnvHelper)}
		makeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrRPC)
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
			handler, request, response := getMethodTypes(fn, rpcModuleId)

			f.Func().Params(Id("resource").Id(resourceName)).Id(handler).
				Params(Id("ctx").Qual("context", "Context"), Id("request").Id(request), Id("response").Id(response)).
				Params(Err().Error()).Block(
				Return(
					Id(handler).
						CallFunc(makeHandlerWrapperCall(fn, info, func(depName string) Code {
							return Id("resource").Dot(depName)
						})).Call(Id("ctx"), Id("request"), Id("response"))))
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

func ifErrorReturnErrRPC(scope *Group, statement *Statement) {
	addIfErrorGuard(scope, statement, Err())
}

func getRPCResourceName(info *PackageInfo) string {
	return "Resource__" + info.Service.Alias
}

func makeHelpersRPC(f *File) {
	addGetEnvHelper(f)
}
