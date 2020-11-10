package dapr

import (
	"github.com/angrypie/tie/template"
	jen "github.com/dave/jennifer/jen"
)

const daprCommon = "github.com/dapr/go-sdk/dapr/proto/common/v1"
const daprRuntime = "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
const pbAny = "github.com/golang/protobuf/ptypes/any"
const pbEmpty = "github.com/golang/protobuf/ptypes/empty"

func genFuncOnInvoke(resource string) jen.Code {
	return jen.Func().Params(
		jen.Id("s").Op("*").Id(resource)).Id("OnInvoke").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("in").Op("*").Qual(daprCommon, "InvokeRequest")).Params(
		jen.Op("*").Qual(daprCommon, "InvokeResponse"),
		jen.Id("error")).Block(
		jen.Var().Id("response").Id("string"),
		jen.Switch(jen.Id("in").Dot("Method")).Block(
			jen.Case(jen.Lit("EchoMethod")).Block()),
		jen.Return().List(jen.Op("&").Qual(daprCommon,
			"InvokeResponse").Values(jen.Id("ContentType").Op(":").Lit("text/plain; charset=UTF-8"),
			jen.Id("Data").Op(":").Op("&").Qual(pbAny, "Any").Values(jen.Id("Value").Op(":").Index().Id("byte").Call(jen.Id("response")))),
			jen.Id("nil")),
	)
}
func genFuncListTopicSubscriptions(resource string) jen.Code {
	return jen.Func().Params(
		jen.Id("s").Op("*").Id(resource)).Id("ListTopicSubscriptions").Params(
		jen.Id("ctx").Qual("context",
			"Context"),
		jen.Id("in").Op("*").Qual(pbEmpty, "Empty")).Params(
		jen.Op("*").Qual(daprRuntime, "ListTopicSubscriptionsResponse"),
		jen.Id("error")).Block(
		jen.Return().List(jen.Op("&").Qual(daprRuntime,
			"ListTopicSubscriptionsResponse").Values(jen.Id("Subscriptions").Op(":").Index().Op("*").Qual(daprRuntime,
			"TopicSubscription").Values(jen.Values(jen.Id("Topic").Op(":").Lit("TopicA")))),
			jen.Id("nil")),
	)
}
func genFuncListInputBindings(resource string) jen.Code {
	return jen.Func().Params(
		jen.Id("s").Op("*").Id(resource)).Id("ListInputBindings").Params(
		jen.Id("ctx").Qual("context",
			"Context"),
		jen.Id("in").Op("*").Qual(pbEmpty, "Empty")).Params(
		jen.Op("*").Qual(daprRuntime, "ListInputBindingsResponse"),
		jen.Id("error")).Block(
		jen.Return().List(jen.Op("&").Qual(daprRuntime,
			"ListInputBindingsResponse").Values(jen.Id("Bindings").Op(":").Index().Id("string").Values(jen.Lit("storage"))),
			jen.Id("nil")),
	)
}
func genFuncOnBindingEvent(resource string) jen.Code {
	return jen.Func().Params(
		jen.Id("s").Op("*").Id(resource)).Id("OnBindingEvent").Params(
		jen.Id("ctx").Qual("context",
			"Context"),
		jen.Id("in").Op("*").Qual(daprRuntime, "BindingEventRequest")).Params(
		jen.Op("*").Qual(daprRuntime, "BindingEventResponse"),
		jen.Id("error")).Block(
		jen.Qual("fmt",
			"Println").Call(jen.Lit("Invoked from binding")),
		jen.Return().List(jen.Op("&").Qual(daprRuntime,
			"BindingEventResponse").Values(),
			jen.Id("nil")),
	)
}
func genFuncOnTopicEvent(resource string) jen.Code {
	return jen.Func().Params(
		jen.Id("s").Op("*").Id(resource)).Id("OnTopicEvent").Params(
		jen.Id("ctx").Qual("context",
			"Context"),
		jen.Id("in").Op("*").Qual(daprRuntime,
			"TopicEventRequest")).Params(
		jen.Op("*").Qual(daprRuntime, "TopicEventResponse"),
		jen.Id("error")).Block(
		jen.Qual("fmt",
			"Println").Call(jen.Lit("Topic message arrived")),
		jen.Return().List(jen.Op("&").Qual(daprRuntime, "TopicEventResponse").Values(),
			jen.Id("nil")),
	)
}
func genDaprAppMethods(f *jen.File, resource string) {
	f.Add(genFuncOnInvoke(resource))
	f.Add(genFuncListTopicSubscriptions(resource))
	f.Add(genFuncListInputBindings(resource))
	f.Add(genFuncOnBindingEvent(resource))
	f.Add(genFuncOnTopicEvent(resource))
}

func genInitGrpcServer(g *jen.Group, instance string) {
	g.List(jen.Id("lis"), jen.Err()).Op(":=").
		Qual("net", "Listen").Call(jen.Lit("tcp"), jen.Lit(":50001"))

	template.AddIfErrorGuard(g, nil, "err", jen.Err())

	g.Id("s").Op(":=").Qual(grpcImport, "NewServer").Call()
	g.Qual(pbImport, "RegisterAppCallbackServer").Call(jen.Id("s"), jen.Id(instance))

	serveStmt := jen.Err().Op(":=").Id("s").Dot("Serve").Call(jen.Id("lis"))
	template.AddIfErrorGuard(g, serveStmt, "err", jen.Err())
}
