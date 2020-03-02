package rpcmod

import (
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	. "github.com/dave/jennifer/jen"
)

const clientNamesSuffix = ""
const rpcxClient = "github.com/smallnest/rpcx/client"

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

	addGetRpcClient(info, f)
	template.ForEachFunction(info, true, func(fn *parser.Function) {
		_, request, response := template.GetMethodTypes(fn, clientNamesSuffix)
		rpcMethodName, _, _ := template.GetMethodTypes(fn, rpcModuleId)

		args := fn.Arguments
		body := func(g *Group) {
			g.Id("response").Op(":=").New(Id(response))
			g.Id("request").Op(":=").New(Id(request))

			if len(args) != 0 {
				g.ListFunc(template.CreateArgsListFunc(args, "request")).Op("=").
					ListFunc(template.CreateArgsListFunc(args))
			}

			g.List(Id("xclient"), Id("err")).Op(":=").Id(getRpcClientFnName).Call()
			template.AddIfErrorGuard(g, nil, nil)
			g.Defer().Id("xclient").Dot("Close").Call()

			g.Err().Op("=").Id("xclient").Dot("Call").Call(
				Qual("context", "Background").Call(), Lit(rpcMethodName),
				Id("request"), Id("response"),
			)
			template.AddIfErrorGuard(g, nil, nil)
			g.Return(ListFunc(template.CreateArgsListFunc(fn.Results, "response")))
		}

		f.Func().ListFunc(func(g *Group) {
			if template.HasReceiver(fn) {
				g.Params(Id("resource").Id(fn.Receiver.Type))
				return
			}
		}).Id(fn.Name).
			ParamsFunc(template.CreateSignatureFromArgs(args)).
			ParamsFunc(template.CreateSignatureFromArgs(fn.Results)).BlockFunc(body)
	})
}

const zeroconf = "github.com/grandcat/zeroconf"

const getRpcClientFnName = "getRpcClient"

//addGetRpcClient returns getRpcClient call Statement
func addGetRpcClient(info *PackageInfo, f *File) {
	f.Add(genFuncgetLocalService())

	f.Func().Id(getRpcClientFnName).Params().
		Params(Id("xclient").Qual(rpcxClient, "XClient"), Err().Error()).
		BlockFunc(func(g *Group) {
			resourceName := getResourceName(info)
			g.List(Id("port"), Id("err")).Op(":=").Id("getLocalService").Call(Lit(resourceName))
			template.AddIfErrorGuard(g, List(), List())
			//TODO p2p discovery
			g.Id("addr").Op(":=").Qual("fmt", "Sprintf").Call(Lit("tcp@127.0.0.1:%d"), Id("port"))

			g.Id("d").Op(":=").Qual(rpcxClient, "NewPeer2PeerDiscovery").Call(Id("addr"), Lit(""))
			g.Id("xclient").Op("=").Qual(rpcxClient, "NewXClient").Call(
				Lit(resourceName), Qual(rpcxClient, "Failtry"),
				Qual(rpcxClient, "RandomSelect"), Id("d"),
				Qual(rpcxClient, "DefaultOption"),
			)
			g.Return()
		})
}

func genFuncgetLocalService() Code {
	return Func().Id("getLocalService").Params(Id("service").Id("string")).Params(Id("port").Id("int"), Id("err").Id("error")).Block(List(Id("resolver"), Id("err")).Op(":=").Qual(zeroconf, "NewResolver").Call(Id("nil")), If(Id("err").Op("!=").Id("nil")).Block(Return().List(Id("port"), Id("err"))), Id("entries").Op(":=").Id("make").Call(Chan().Op("*").Qual(zeroconf, "ServiceEntry")), List(Id("ctx"), Id("cancel")).Op(":=").Qual("context", "WithTimeout").Call(Qual("context", "Background").Call(), Qual("time", "Second").Op("*").Lit(5)), Id("err").Op("=").Id("resolver").Dot("Browse").Call(Id("ctx"), Id("service"), Lit("local."), Id("entries")), If(Id("err").Op("!=").Id("nil")).Block(Return().List(Id("port"), Id("err"))), Select().Block(Case(Op("<-").Id("ctx").Dot("Done").Call()).Block(Id("cancel").Call(), Return().List(Id("port"), Qual("errors", "New").Call(Lit("Service not found")))), Case(Id("entry").Op(":=").Op("<-").Id("entries")).Block(Id("cancel").Call(), Return().List(Id("entry").Dot("Port"), Id("nil")))))
}
