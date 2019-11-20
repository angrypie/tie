package template

import (
	"fmt"
	"regexp"
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
			jsonTag := strings.ToLower(name)
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

	forEachFunction(info.Functions, true, func(fn *parser.Function) {
		_, reqName, respName := getMethodTypes(fn, postfix)
		code.Add(createTypeFromArgs(reqName, fn.Arguments, info))
		code.Line()
		code.Add(createTypeFromArgs(respName, fn.Results, info))
		code.Line()
	})
	return code
}

func getMethodTypes(fn *parser.Function, postfix string) (handler, request, response string) {
	method, receiver := fn.Name, fn.Receiver.Type
	handler = fmt.Sprintf("%s%s%sHandler", receiver, method, postfix)
	request = fmt.Sprintf("%s%s%sRequest", receiver, method, postfix)
	response = fmt.Sprintf("%s%s%sResponse", receiver, method, postfix)
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

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	return strings.ToLower(
		matchAllCap.ReplaceAllString(str, "${1}_${2}"),
	)
}

func getReceiverVarName(receiverTypeName string) string {
	if receiverTypeName == "" {
		return ""
	}
	return fmt.Sprintf("Receiver%s", receiverTypeName)
}

func hasReceiver(fn *parser.Function) bool {
	return fn.Receiver.Type != ""
}

func isTopLevelInitReceiver(fn *parser.Function) bool {
	if fn == nil {
		return false
	}
	for _, field := range fn.Arguments {
		name := field.Name
		if name != "getEnv" {
			return false
		}
	}
	return true
}

func getInitReceiverDepsNames(fn *parser.Function) (code Code) {
	if fn == nil {
		return
	}
	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			injectReceiverName(g, field)
		}
	})
}

var matchFuncType = regexp.MustCompile("^func.*")

//injectReceiverName injects recevier variable name to given scope.
func injectReceiverName(g *Group, field parser.Field) {
	t := field.Type
	if matchFuncType.MatchString(t) {
		return
	}
	depVarName := getReceiverVarName(t)
	g.Id(depVarName)
}

func getInitReceiverDepsSignature(fn *parser.Function, info *PackageInfo) (code Code) {
	if fn == nil {
		return
	}
	return ListFunc(func(g *Group) {
		for _, field := range fn.Arguments {
			t := field.Type
			if matchFuncType.MatchString(t) {
				return
			}
			depVarName := getReceiverVarName(t)
			g.Id(depVarName).Op("*").Qual(info.Service.Name, strings.Trim(t, "*"))
		}
	})
}

func injectOriginalMethodCall(g *Group, fn *parser.Function, method Code) {
	g.ListFunc(createArgsListFunc(fn.Results, "response")).
		Op("=").Add(method).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
}
