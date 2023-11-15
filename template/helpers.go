package template

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
)

var uniqueNames map[string]string

func init() {
	uniqueNames = make(map[string]string)
}

func ID(name ...string) string {
	prefix := "id"
	if len(name) != 0 {
		prefix = strings.Join(name, "")
		if value, ok := uniqueNames[prefix]; ok {
			return value
		}
	}
	id := fmt.Sprintf("%s__%d", prefix, rand.Intn(999999))
	uniqueNames[prefix] = id
	return id
}

//TODO move to PackageInfo
func GetMethodTypes(fn parser.Function) (handler, request, response string) {
	method, receiver := fn.Name, ""
	if HasReceiver(fn) {
		receiver = fn.Receiver.TypeName()
	}

	handler = ID(receiver, method, "Handler")
	request = ID(receiver, method, "Request")
	response = ID(receiver, method, "Response")
	return
}

func GetReceiverVarName(receiverTypeName string) string {
	if receiverTypeName == "" {
		return ""
	}
	return fmt.Sprintf("Receiver%s", receiverTypeName)
}

func HasReceiver(fn parser.Function) bool {
	return fn.Receiver.IsDefined()
}

//HasTopLevelReceiver returns false if function has other receiver as argumenet.
func HasTopLevelReceiver(fn parser.Function, info *PackageInfo) bool {
	for _, field := range fn.Arguments {
		if _, ok := info.GetConstructor(field); ok {
			return false
		}
	}
	return true
}

//ForEachFunction executes callback for each function in package
//expecept special Stop function that shoudn't be externaly exposed.
func ForEachFunction(info *PackageInfo, skipInit bool, cb func(parser.Function)) {
	fns := info.Functions
	if skipInit {
		fns = getFnsWithoutConstructors(info)
	}
	for _, fn := range fns {
		if fn.Name == "Stop" {
			continue
		}
		cb(fn)
	}

}

//getFnsWithoutConstructors filters constructors from info.Functions
func getFnsWithoutConstructors(info *PackageInfo) (filtered []parser.Function) {
	for _, fn := range info.Functions {
		if _, ok := isConventionalConstructor(fn); !ok {
			filtered = append(filtered, fn)
		}
	}
	return
}

var getTypeFromConstructorName = regexp.MustCompile(`\ANew(.*)\z`)

func isConventionalConstructor(fn parser.Function) (receiver parser.Field, ok bool) {
	if HasReceiver(fn) {
		return
	}

	match := getTypeFromConstructorName.FindStringSubmatch(fn.Name)
	if len(match) < 2 {
		return
	}
	recType := match[1]

	for _, ret := range fn.Results.List() {
		if ret.TypeName() == recType {
			return ret, true
		}
	}

	return
}

func TrimPrefix(str string) string {
	return strings.TrimPrefix(str, "*")
}

var matchFuncType = regexp.MustCompile("^func.*")

func isFuncType(t string) bool {
	return matchFuncType.MatchString(t)
}

//filterNotReceiverArgs removes receivers args from filed list
func filterHelperArgs(fields []parser.Field, info *PackageInfo) (filtered []parser.Field) {
	for _, field := range fields {
		if cons, ok := info.GetConstructor(field); ok && HasTopLevelReceiver(cons.Function, info) {
			continue
		}
		filtered = append(filtered, field)
	}
	return
}

type ForEachRecCb = func(receiver parser.Field, constructor OptionalConstructor)

//TODO save receivers during PackageInfo initialization
//MakeForEachReceiver executes callback for each receiver.
func MakeForEachReceiver(
	info *PackageInfo, cb ForEachRecCb,
) (receiversProcessed map[string]parser.Field) {
	receiversProcessed = make(map[string]parser.Field)
	cbWrapper := func(receiver parser.Field, constructor OptionalConstructor) {
		receiversProcessed[receiver.TypeName()] = receiver
		cb(receiver, constructor)
	}
	//Create receivers for each constructor
	for _, c := range info.Constructors {
		cbWrapper(c.Receiver, NewOptionalConstructor(c))
	}

	//Create receivers that does not have constructor
	ForEachFunction(info, false, func(fn parser.Function) {
		//Skip function if it does not have receiver
		if !HasReceiver(fn) {
			return
		}
		receiver := fn.Receiver
		receiverType := receiver.TypeName()
		// Skip if receiver already created.
		if _, ok := receiversProcessed[receiverType]; ok {
			return
		}
		//It will not create constructor call due constructor func is nil
		cbWrapper(receiver, NewOptionalConstructor())
	})

	return receiversProcessed
}

func ReqRecName(fn parser.Function) string {
	return strings.Title(fn.Receiver.Name())
}

//TODO return struct type

func CreateCombinedHandlerArgs(fn parser.Function, info *PackageInfo) (fields []types.Field) {
	fields = fieldsFromParser(fn.Arguments)
	if !HasReceiver(fn) {
		return
	}
	cons, ok := info.GetConstructor(fn.Receiver)
	if ok && !HasTopLevelReceiver(cons.Function, info) {
		fields = append(fields, NewField(ReqRecName(fn), fn.Receiver.TypeName()))
	}

	return
}

func GetResourceName(info *PackageInfo) string {
	return "Resource__" + info.PackageName
}
