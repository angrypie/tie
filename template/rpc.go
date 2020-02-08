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
		makeWaitGuard(g)
	})

	return fmt.Sprintf("%#v", f), nil
}

func makeRPCServer(info *PackageInfo, main *Group, f *File) {
	service := info.Service
	if service.Type == "httpOnly" {
		return
	}

	makeRPCRequestResponseTypes(info, main, f)
	makeRPCHandlers(info, main, f)

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
				makeRPCReceiverMiddleware(receiverVarName, g, constructorFunc, info) //RC
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
	if hasReceiver(fn) {
		file.Func().Id(handler).ParamsFunc(func(g *Group) {
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
	} else {
		file.Func().Id(handler).
			Params(Id("ctx").Qual("context", "Context"), Id("request").Id(request), Id("response").Id(response)). //RC
			Params(Err().Error()).BlockFunc(handlerBody).
			Line()
	}

}

func ifErrorReturnErrRPC(scope *Group, statement *Statement) {
	scope.If(
		statement,
		Err().Op("!=").Nil(),
	).Block(
		Return(Err()),
	)
}

func makeRPCReceiverMiddleware(recId string, scope *Group, constructor *parser.Function, info *PackageInfo) {
	if constructor == nil {
		return
	}
	constructorCall := func(g *Group) {
		for _, field := range constructor.Arguments {
			name := field.Name
			//TODO check and getEnv function signature
			//Inject getEnv function that provide access to environment variables
			//RC deleted getHeader
			if name == "getEnv" {
				g.Id(getEnvHelper)
				continue
			}

			//TODO send nil for pointer or empty object otherwise
			if !info.IsReceiverType(field.Type) {
				//g.Id("request").Dot(field.Name)
				g.ListFunc(createArgsListFunc([]parser.Field{field}, "request"))
				continue
			}

			//Oterwise inject receiver dependencie
			g.Id(getReceiverVarName(field.Type))
		}
	}

	ifErrorReturnErrRPC(
		scope,
		List(Id(recId), Err()).Op("=").Qual(info.Service.Name, constructor.Name).CallFunc(constructorCall),
	)
}
