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

func forEachFunction(fns []*parser.Function, skipInit bool, cb func(*parser.Function)) {
	if skipInit {
		fns = removeConstructors(fns)
	}
	for _, fn := range fns {
		cb(fn)
	}

}

//removeConstructors removes type constructors
func removeConstructors(fns []*parser.Function) (filtered []*parser.Function) {
	processed := map[string]bool{"": true}
	inits := make(map[*parser.Function]bool)
	for _, fn := range fns {
		//Do not search contructor for already processed receiver.
		if processed[fn.Receiver.Type] {
			continue
		}
		//Save function that is contructor for type.
		processed[fn.Receiver.Type], inits[findInitReceiver(fns, fn)] = true, true
	}

	for _, fn := range fns {
		if !inits[fn] {
			filtered = append(filtered, fn)
		}
	}
	return
}

//createIsConstructor returns function that checks if function name is type contructor.
func createIsConstructor(typeName string) func(funcName string) bool {
	reg := fmt.Sprintf(`\ANew%s\z`, typeName)
	return func(funcName string) bool {
		match, _ := regexp.MatchString(reg, funcName)
		return match
	}
}

//findInitReceiver find conventional contructor for function receiver type.
func findInitReceiver(fns []*parser.Function, forFunc *parser.Function) *parser.Function {
	isContsructor := createIsConstructor(forFunc.Receiver.Type)

	for _, fn := range fns {
		if !hasReceiver(fn) && isContsructor(fn.Name) {
			return fn
		}
	}
	return nil
}
