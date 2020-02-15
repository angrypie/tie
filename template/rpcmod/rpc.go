package rpcmod

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	. "github.com/dave/jennifer/jen"
)

const rpcxServer = "github.com/smallnest/rpcx/server"
const rpcModuleId = "RPC"

type PackageInfo = template.PackageInfo

func GetModule(info *PackageInfo) (string, error) {
	f := NewFile(strings.ToLower(rpcModuleId))

	f.Func().Id("Main").Params().BlockFunc(func(main *Group) {
		template.MakeGracefulShutdown(info, main, f)
		template.MakeInitService(info, main)

		makeStartRPCServer(info, main, f)
	})

	template.MakeHandlers(info, f, makeRPCHandler)
	f.Add(template.CreateReqRespTypes(rpcModuleId, info))
	template.AddGetEnvHelper(f)

	return fmt.Sprintf("%#v", f), nil
}

func makeRPCHandler(info *PackageInfo, fn *parser.Function, file *Group) {

	handlerBody := func(g *Group) {
		middlewares := template.MiddlewaresMap{"getEnv": Id(template.GetEnvHelper)}
		template.MakeOriginalCall(info, fn, g, middlewares, ifErrorReturnErrRPC)
		g.Return(Nil())
	}

	_, request, response := template.GetMethodTypes(fn, rpcModuleId)

	template.MakeHandlerWrapper(
		rpcModuleId, handlerBody, info, fn, file,
		List(Id("ctx").Qual("context", "Context"), Id("request").Id(request), Id("response").Id(response)),
		Err().Error(),
	)
}

func makeStartRPCServer(info *PackageInfo, main *Group, f *File) {
	main.Go().Id("startServer").Call()

	f.Func().Id("startServer").Params().BlockFunc(func(g *Group) {
		template.MakeStartServerInit(info, g) //SIM
		receiversCreated := template.MakeReceiversForHandlers(info, g)

		//RC replace http server init
		resourceName := "Resource__" + info.Service.Alias
		resourceInstance := "Instance__" + resourceName

		f.Type().Id(resourceName).StructFunc(func(g *Group) {
			for receiverType := range receiversCreated {
				receiverVarName := template.GetReceiverVarName(receiverType)
				g.Id(receiverVarName).Op("*").Qual(info.Service.Name, template.TrimPrefix(receiverType))
			}
		})

		//.2 Add handler for each function.
		template.ForEachFunction(info, true, func(fn *parser.Function) {
			handler, request, response := template.GetMethodTypes(fn, rpcModuleId)

			f.Func().Params(Id("resource").Id(resourceName)).Id(handler).
				Params(Id("ctx").Qual("context", "Context"), Id("request").Id(request), Id("response").Id(response)).
				Params(Err().Error()).Block(
				Return(
					Id(handler).
						CallFunc(template.MakeHandlerWrapperCall(fn, info, func(depName string) Code {
							return Id("resource").Dot(depName)
						})).Call(Id("ctx"), Id("request"), Id("response"))))
		})

		g.Id("server").Op(":=").Qual(rpcxServer, "NewServer").Call()
		g.Id(resourceInstance).Op(":=").Op("&").Id(resourceName).Values(DictFunc(func(d Dict) {
			for receiverType := range receiversCreated {
				receiverVarName := template.GetReceiverVarName(receiverType)
				d[Id(receiverVarName)] = Id(receiverVarName)
			}
		}))
		g.Id("server").Dot("RegisterName").Call(Lit(resourceName), Id(resourceInstance), Lit(""))

		g.Id("server").Dot("Serve").Call(Lit("tcp"), Id("address"))

		//RC end
	})
}

func ifErrorReturnErrRPC(scope *Group, statement *Statement) {
	template.AddIfErrorGuard(scope, statement, Err())
}
