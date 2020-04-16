package template

import (
	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

//TODO consider to remove moduld in naming
func MakeStartRPCServer(
	info *PackageInfo, moduleId string, main *Group, f *File,
	cb func(g *Group, resource, instance string),
) {
	//TODO handle thi error
	main.Err().Op(":=").Id("startServer").Call()
	main.If(Err().Op("!=").Nil()).Block(Panic(Err()))

	f.Func().Id("startServer").Params().Params(Err().Error()).BlockFunc(func(g *Group) {
		receiversCreated := MakeReceiversForHandlers(info, g)

		resourceName := GetResourceName(info)
		resourceInstance := "Instance__" + resourceName

		f.Type().Id(resourceName).StructFunc(func(g *Group) {
			for receiverType := range receiversCreated {
				receiverVarName := GetReceiverVarName(receiverType)
				g.Id(receiverVarName).Op("*").Qual(info.GetServicePath(), TrimPrefix(receiverType))
			}
		})

		//.2 Add handler for each function.
		ForEachFunction(info, true, func(fn *parser.Function) {
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
