package rpcmod

import (
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
	f := NewFile(strings.ToLower(rpcModuleId))

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

	resourceName := getResourceName(info)

	template.ForEachFunction(info, true, func(fn *parser.Function) {
		_, request, response := template.GetMethodTypes(fn, clientNamesSuffix)
		rpcMethodName, _, _ := template.GetMethodTypes(fn, rpcModuleId)

		body := func(g *Group) {
			g.Id("request").Op(":=").New(Id(request))
			g.Id("response").Op(":=").New(Id(response))

			g.Err().Op("=").Id("client").Dot("Call").Call(
				Qual("context", "Background").Call(),
				Lit(resourceName), Lit(rpcMethodName),
				Id("request"), Id("response"),
			)
			g.Return()
		}

		f.Func().ListFunc(func(g *Group) {
			if template.HasReceiver(fn) {
				g.Params(Id("resource").Id(fn.Receiver.Type))
				return
			}
		}).Id(fn.Name).
			ParamsFunc(template.CreateSignatureFromArgs(fn.Arguments)).
			ParamsFunc(template.CreateSignatureFromArgs(fn.Results)).BlockFunc(body)
	})
}
