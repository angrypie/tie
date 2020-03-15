package micromod

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	. "github.com/dave/jennifer/jen"
)

const clientNamesSuffix = ""

func NewClientModule(p *parser.Parser) template.Module {
	return template.NewStandartModule("client", GenerateClient, p, nil)
}

func GenerateClient(p *parser.Parser) (pkg *template.Package) {
	info := template.NewPackageInfoFromParser(p)
	//TODO all modules noods to create upgraded subpackage, to make ServicePath reusable
	info.ServicePath = info.Service.Name + "/tie_modules/micromod/upgraded"
	f := NewFile(strings.ToLower(microModuleId))

	f.Add(template.CreateReqRespTypes(clientNamesSuffix, info))

	makeClientAPI(info, f)

	return &template.Package{
		Name:  "client",
		Files: [][]byte{[]byte(f.GoString())},
	}
}

func makeClientAPI(info *PackageInfo, f *File) {

	cb := func(receiverType string, constructor *parser.Function) {
		f.Type().Id(receiverType).Struct()
	}
	template.MakeForEachReceiver(info, cb)

	template.ForEachFunction(info, true, func(fn *parser.Function) {
		_, request, response := template.GetMethodTypes(fn, clientNamesSuffix)
		rpcMethodName, _, _ := template.GetMethodTypes(fn, microModuleId)
		resourceName := template.GetResourceName(info)

		args := fn.Arguments
		body := func(g *Group) {
			g.Id("response").Op(":=").New(Id(response))
			g.Id("request").Op(":=").New(Id(request))

			if len(args) != 0 {
				g.ListFunc(template.CreateArgsListFunc(args, "request")).Op("=").
					ListFunc(template.CreateArgsListFunc(args))
			}

			g.Id("service").Op(":=").Qual(gomicro, "NewService").Call()
			g.Id("service").Dot("Init").Call()
			g.Id("c").Op(":=").Id("service").Dot("Client").Call()

			g.Id("microRequest").Op(":=").Id("client").Dot("NewRequest").Call(
				Lit(resourceName), Lit(fmt.Sprintf("%s.%s", resourceName, rpcMethodName)),
				Id("request"),
				Qual(gomicroClient, "WithContentType").Call(Lit("application/json")),
			)

			g.Err().Op("=").Id("c").Dot("Call").Call(
				Qual("context", "TODO").Call(), Id("microRequest"), Id("response"))
			template.AddIfErrorGuard(g, nil, nil)

			g.Return(ListFunc(template.CreateArgsListFunc(fn.Results, "response")))
		}

		f.Func().ListFunc(func(g *Group) {
			if template.HasReceiver(fn) {
				g.Params(Id("resource").Id(fn.Receiver.Type))
				return
			}
		}).Id(fn.Name).
			ParamsFunc(template.CreateSignatureFromArgs(args, info)).
			ParamsFunc(template.CreateSignatureFromArgs(fn.Results, info)).BlockFunc(body)
	})
}
