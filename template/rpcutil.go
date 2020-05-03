package template

import (
	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

func MakeStartRPCServer(
	info *PackageInfo, main *Group, f *File,
	cb func(g *Group, resource, instance string),
) {
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
}

func GetRpcHandlerArgsList(request, response string) *Statement {
	return List(
		Id("ctx").Qual("context", "Context"),
		Id("request").Op("*").Id(request),
		Id("response").Op("*").Id(response),
	)
}

func GetResourceName(info *PackageInfo) string {
	return "Resource__" + info.PackageName
}

type ClientMethodBody = func(ids ClientMethodIds) *Statement

type ClientMethodIds struct {
	Request, Response, Method, Resource, Err string
}

func ClientMethods(info *PackageInfo, body ClientMethodBody) *Statement {
	stmt := Comment("Client Methods").Line()
	ForEachFunction(info, true, func(fn parser.Function) {
		stmt.Add(ClientMethod(fn, info, body))
	})

	return stmt
}

func ClientMethod(fn parser.Function, info *PackageInfo, body ClientMethodBody) *Statement {
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
		g.Add(body(ClientMethodIds{
			Method:   rpcMethodName,
			Resource: resourceName,
			Err:      errId,
			Request:  request,
			Response: response,
		}))

		AddIfErrorGuard(g, AssignErrToResults(Id(errId), fn.Results), errId, nil)

		g.Return(ListFunc(CreateArgsListFunc(fn.Results.List(), response)))
	}

	return Func().ListFunc(func(g *Group) {
		if HasReceiver(fn) {
			g.Params(Id("resource").Id(fn.Receiver.TypeName()))
			return
		}
	}).Id(fn.Name).
		ParamsFunc(CreateSignatureFromArgs(args, info)).
		ParamsFunc(CreateSignatureFromArgs(fn.Results.List(), info)).
		BlockFunc(baseBody).Line()
}

func TemplateClient(info *PackageInfo, body ClientMethodBody) *Statement {
	code := Comment("Naive client template").Line()
	code.Add(
		CreateReqRespTypes(info),
		CreateTypeAliases(info),
		ClientMethods(info, body),
	)
	return code
}
