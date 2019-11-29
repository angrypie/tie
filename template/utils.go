package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
)

func getMethodTypes(fn *parser.Function, postfix string) (handler, request, response string) {
	method, receiver := fn.Name, fn.Receiver.Type
	handler = fmt.Sprintf("%s%s%sHandler", receiver, method, postfix)
	request = fmt.Sprintf("%s%s%sRequest", receiver, method, postfix)
	response = fmt.Sprintf("%s%s%sResponse", receiver, method, postfix)
	return
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

//hasTopLevelReceiver returns true if construcotor has other receiver as argumenet.
func hasTopLevelReceiver(fn *parser.Function, info *PackageInfo) bool {
	if fn == nil {
		return false
	}
	for _, field := range fn.Arguments {
		if _, ok := info.Constructors[field.Type]; ok {
			return false
		}
	}
	return true
}

func forEachFunction(info *PackageInfo, skipInit bool, cb func(*parser.Function)) {
	fns := info.Functions
	if skipInit {
		fns = getFnsWithoutConstructors(info)
	}
	for _, fn := range fns {
		cb(fn)
	}

}

//getFnsWithoutConstructors removes type constructors
func getFnsWithoutConstructors(info *PackageInfo) (filtered []*parser.Function) {
	fns := info.Functions

	//Get all constructors
	constructors := make(map[*parser.Function]bool)
	for _, fn := range info.Constructors {
		constructors[fn] = true
	}

	for _, fn := range fns {
		if !constructors[fn] {
			filtered = append(filtered, fn)
		}
	}
	return
}

var getTypeFromConstructorName = regexp.MustCompile(`\ANew(.*)\z`)

func isConventionalConstructor(fn *parser.Function) (ok bool, _type string) {
	if hasReceiver(fn) {
		return
	}

	rets := make(map[string]bool)
	for _, ret := range fn.Results {
		rets[ret.Type] = true
	}
	match := getTypeFromConstructorName.FindStringSubmatch(fn.Name)
	if len(match) < 2 {
		return
	}

	return rets[match[1]], match[1]
}
