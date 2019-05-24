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

func createArgsListFunc(args []parser.Field, params ...string) func(*Group) {
	return createArgsList(args, func(arg *Statement) *Statement {
		return arg
	}, params...)
}

func createArgsList(args []parser.Field, transform func(*Statement) *Statement, params ...string) func(*Group) {
	prefix, typeNames := "", ""
	if len(params) > 0 {
		prefix = params[0]
	}
	if len(params) > 1 {
		typeNames = params[1]
	}
	return func(g *Group) {
		for _, arg := range args {
			//Skip iteration if argument type not specified
			if typeNames != "" && !strings.Contains(typeNames, arg.Type+",") {
				continue
			}
			if isArgNameAreDTO(arg.Name) && prefix != "" {
				g.Add(transform(Id(prefix).Dot(arg.Type)))
				return
			}
			name := strings.Title(arg.Name)
			if prefix != "" {
				g.Add(transform(Id(prefix).Dot(name)))
			} else {
				g.Add(transform(Id(name)))
			}
		}
	}
}

func isArgNameAreDTO(name string) bool {
	n := strings.ToLower(name)
	return n == "requestdto" || n == "responsedto"
}
