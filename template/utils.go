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
			if isArgNameAreDTO(name) {
				name = ""
			}
			field := Id(strings.Title(name)).Op(arg.Prefix)
			if arg.Package != "" {
				field.Qual(info.Service.Name, arg.Type)
			} else {
				field.Id(arg.Type)
			}
			jsonTag := fmt.Sprintf("%s,omitempty", strings.ToLower(name))
			if arg.Type == "error" {
				jsonTag = "-"
			}
			field.Tag(map[string]string{"json": jsonTag})
			g.Add(field)
		}
	})
}

func createReqRespTypes(postfix string, info *PackageInfo) Code {
	code := Comment(fmt.Sprintf("Request/Response types (%s)", postfix)).Line()

	for _, fn := range info.Functions {
		_, reqName, respName := getMethodTypes(fn.Name, postfix)
		code.Add(createTypeFromArgs(reqName, fn.Arguments, info))
		code.Line()
		code.Add(createTypeFromArgs(respName, fn.Results, info))
		code.Line()
	}
	return code
}

func getMethodTypes(method, postfix string) (handler, request, response string) {
	handler = fmt.Sprintf("%s%sHandler", method, postfix)
	request = fmt.Sprintf("%s%sRequest", method, postfix)
	response = fmt.Sprintf("%s%sResponse", method, postfix)
	return
}

func createArgsListFunc(args []parser.Field, prefix ...string) func(*Group) {
	p := ""
	if len(prefix) > 0 {
		p = prefix[0]
	}
	return func(g *Group) {
		for i, arg := range args {
			if i == 0 && isArgNameAreDTO(arg.Name) && p != "" {
				g.Id(p).Dot(arg.Type)
				return
			}
			name := strings.Title(arg.Name)
			if p != "" {
				g.Id(p).Dot(name)
			} else {
				g.Id(name)
			}
		}
	}
}

func isArgNameAreDTO(name string) bool {
	n := strings.ToLower(name)
	return n == "requestdto" || n == "responsedto"
}
