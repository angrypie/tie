package template

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

func createTypeFromArgs(name string, args []parser.Field, info *PackageInfo) Code {
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, arg := range args {
			name := arg.Name
			if name == "requestDTO" || name == "responseDTO" {
				name = ""
			}
			field := Id(strings.Title(name)).Op(arg.Prefix)
			if arg.Package != "" {
				field.Qual(info.Path, arg.Type)
			} else {
				field.Id(arg.Type)
			}
			g.Add(field)
		}
	})
}

func createReqRespTypes(postfix string, info *PackageInfo) Code {
	code := Comment(fmt.Sprintf("Request/Response types (%s)", postfix)).Line()

	for _, fn := range info.Functions {
		respName := fmt.Sprintf("%sResponse%s", fn.Name, postfix)
		reqName := fmt.Sprintf("%sRequest%s", fn.Name, postfix)
		code.Add(createTypeFromArgs(reqName, fn.Arguments, info))
		code.Line()
		code.Add(createTypeFromArgs(respName, fn.Results, info))
	}
	return code
}
