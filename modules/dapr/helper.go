package dapr

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	. "github.com/dave/jennifer/jen"
)

const daprCommon = "github.com/dapr/go-sdk/dapr/proto/common/v1"
const daprRuntime = "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
const pbAny = "github.com/golang/protobuf/ptypes/any"
const pbEmpty = "github.com/golang/protobuf/ptypes/empty"
const daprdImport = "github.com/dapr/go-sdk/service/grpc"

const serverInstance = "DaprService"

func genMethodHandlers(info *template.PackageInfo, g *Group, f *File) {
	//.2 Add handler for each function.
	template.ForEachFunction(info, true, func(fn parser.Function) {
		handler, _, _ := template.GetMethodTypes(fn)

		route := fmt.Sprintf("/%s", fn.Name)
		if fn.Receiver.IsDefined() {
			route = fmt.Sprintf("%s/%s", fn.Receiver.TypeName(), fn.Name)
		}

		g.Id(serverInstance).Dot("AddServiceInvocationHandler").Call(
			Lit(toSnakeCase(route)),
			Id(handler).CallFunc(template.MakeHandlerWrapperCall(fn, info, func(depName string) Code {
				return Id(depName)
			})),
		)
	})
}

func genInitGrpcServer(g *Group, instance string) {

	g.List(Id(serverInstance), Err()).Op(":=").Qual(daprdImport, "NewService").Call(Lit(":50001"))
	template.AddIfErrorGuard(g, nil, "err", Err())

	startStmt := Err().Op(":=").Id(serverInstance).Dot("Start").Call()
	template.AddIfErrorGuard(g, startStmt, "err", Err())
}

func ifErrorOnInvoke(scope *Group, statement *Statement) {
}

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	return strings.ToLower(
		matchAllCap.ReplaceAllString(str, "${1}_${2}"),
	)
}
