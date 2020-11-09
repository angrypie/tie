package template

import (
	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

//StartRPCServerCb is used to insert specific code to server init template.
type StartRPCServerCb = func(g *Group, resource, instance string)

//MakeStartRPCServer creates server initialization method.
func MakeStartRPCServer(info *PackageInfo, cb StartRPCServerCb, main *Group, f *File) {
	main.Comment("MakeStartRPCServer (local)").Line()
	f.Comment("MakeStartRPCServer (file)").Line()

	//TODO handle thi error
	main.Err().Op(":=").Id("startServer").Call()
	main.If(Err().Op("!=").Nil()).Block(Panic(Err()))

	f.Func().Id("startServer").Params().Params(Err().Error()).BlockFunc(func(g *Group) {
		receiversCreated := MakeReceiversForHandlers(info, g)

		resourceName := GetResourceName(info)
		resourceInstance := "Instance___" + resourceName

		f.Type().Id(resourceName).StructFunc(func(g *Group) {
			for receiverType := range receiversCreated {
				receiverVarName := GetReceiverVarName(receiverType)
				g.Id(receiverVarName).Op("*").Qual(info.GetServicePath(), TrimPrefix(receiverType))
			}
		})

		//.2 Add handler for each function.
		ForEachFunction(info, true, func(fn parser.Function) {
			handler, request, response := GetMethodTypes(fn)

			f.Func().Params(Id("resource").Id(resourceName)).Id(handler).
				Params(GetRpcHandlerArgsList(request, response)).
				Params(Err().Error()).Block(
				Return(
					Id(handler).
						CallFunc(MakeHandlerWrapperCall(fn, info, func(depName string) Code {
							return Id("resource").Dot(depName)
						})).Call(Id("ctx"), Id("request"), Id("response"))))
		})

		g.Id(resourceInstance).Op(":=").Op("&").Id(resourceName).Values(DictFunc(func(d Dict) {
			for receiverType := range receiversCreated {
				receiverVarName := GetReceiverVarName(receiverType)
				d[Id(receiverVarName)] = Id(receiverVarName)
			}
		}))
		cb(g, resourceName, resourceInstance)
		g.Return()
	})
	return
}

//GetRpcHandlerArgsList creates args list for rpc handler wrapper.
func GetRpcHandlerArgsList(request, response string) *Statement {
	return List(
		Id("ctx").Qual("context", "Context"),
		Id("request").Op("*").Id(request),
		Id("response").Op("*").Id(response),
	)
}

//ServerMethods creates server methods for each service function.
func ServerMethods(info *PackageInfo, f *File) {
	f.Comment("Server Methods").Line()
	ForEachFunction(info, true, func(fn parser.Function) {
		body := func(g *Group) {
			middlewares := MiddlewaresMap{"getEnv": Id(GetEnvHelper)}
			MakeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrRPC())
			g.Return(Nil())
		}

		_, request, response := GetMethodTypes(fn)
		MakeHandlerWrapper(
			f, body, info, fn,
			GetRpcHandlerArgsList(request, response),
			Err().Error(),
		)
	})
}

//TemplateServer creates template module for RPC server.
func TemplateServer(info *PackageInfo, f *File, cb StartRPCServerCb) {
	f.Func().Id("Main").Params().BlockFunc(func(main *Group) {
		GracefulShutdown(info, main, f)
		MakeInitService(info, main)
		MakeStartRPCServer(info, cb, main, f)
	})

	ServerMethods(info, f)
	CreateReqRespTypes(info, f)
	AddGetEnvHelper(f)
}

//TemplateServer creates template module for RPC client.
func TemplateClient(info *PackageInfo, f *File, body ClientMethodBody) {
	CreateReqRespTypes(info, f)
	CreateTypeAliases(info, f)
	clientMethods(info, body, f)
}

//clientMethods creates client method for each service function.
func clientMethods(info *PackageInfo, body ClientMethodBody, f *File) {
	f.Comment("Client Methods").Line()
	ForEachFunction(info, true, func(fn parser.Function) {
		ClientMethod(fn, info, body, f)
	})
}

//ClientMethodBody is used to insert specific code to client method template.
type ClientMethodBody = func(ids ClientMethodIds, g *Group)

//ClientMethodIds contains identifiers that available in client method template.
type ClientMethodIds struct {
	Request  string //Request variable identifier
	Response string //Response valiable identifier
	Method   string //RPC Method string
	Resource string //RPC Resource string
	Err      string //Error variable identifer
}

//ClientMethod creates client method for given function.
func ClientMethod(fn parser.Function, info *PackageInfo, body ClientMethodBody, f *File) {
	args := fn.Arguments

	baseBody := func(g *Group) {
		rpcMethodName, requestType, responseType := GetMethodTypes(fn)
		request, response := ID("request"), ID("response")

		g.Id(response).Op(":=").New(Id(responseType))
		g.Id(request).Op(":=").New(Id(requestType))

		//Bind method args data to request
		if len(args) != 0 {
			g.ListFunc(CreateArgsListFunc(args, request)).Op("=").
				ListFunc(CreateArgsListFunc(args))
		}
		//Bind receiver data to request
		if HasReceiver(fn) {
			constructor, ok := info.GetConstructor(fn.Receiver)
			if ok && !HasTopLevelReceiver(constructor.Function, info) {
				g.Id(request).Dot(RequestReceiverKey).Op("=").Id("resource")
			}
		}

		resourceName := GetResourceName(info)
		errId := ID("err")
		g.Var().Id(errId).Error()
		g.Id("_").Op("=").Id(errId)

		//Add user body
		body(ClientMethodIds{
			Method:   rpcMethodName,
			Resource: resourceName,
			Err:      errId,
			Request:  request,
			Response: response,
		}, g)

		AddIfErrorGuard(g, AssignErrToResults(Id(errId), fn.Results), errId, nil)

		g.Return(ListFunc(CreateArgsListFunc(fn.Results.List(), response)))
	}

	f.Func().ListFunc(func(g *Group) {
		if HasReceiver(fn) {
			g.Params(Id("resource").Id(fn.Receiver.TypeName()))
			return
		}
	}).Id(fn.Name).
		ParamsFunc(CreateSignatureFromArgs(args, info)).
		ParamsFunc(CreateSignatureFromArgs(fn.Results.List(), info)).
		BlockFunc(baseBody).Line()
}

//ifErrorReturnErrRPC creates error guard for RPC wrapper function.
func ifErrorReturnErrRPC() IfErrorGuard {
	return func(scope *Group, statement *Statement) {
		AddIfErrorGuard(
			scope,
			statement,
			"err",
			Err(),
		)
	}
}
