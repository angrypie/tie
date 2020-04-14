package parser

import (
	"go/build"
	"go/types"
	"log"
	"reflect"
	"strings"
)

type Function struct {
	Name        string
	Arguments   []Field
	Results     []Field
	Receiver    Field
	Package     string
	ServiceType string
}

type TypeSpec struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Var  *types.Var
	Type
}

//IsDefined return true if Var type is set.
func (field Field) IsDefined() bool {
	return field.Var != nil
}

type Type struct {
	typ types.Type
}

func (t Type) TypeString() string {
	return t.typ.String()
}

func (t Type) GetLocalTypeName() string {
	arr := strings.Split(t.TypeString(), t.fullPkgPath()+".")
	return arr[len(arr)-1]
}

func (t Type) PkgPath() string {
	return strings.TrimPrefix(
		t.fullPkgPath(),
		build.Default.GOPATH+"/src/",
	)
}

func (t Type) fullPkgPath() string {
	return traverseType(t.typ)
}

func (t Type) GetTypeParts() (prefix, path, local string) {
	fullPath := t.fullPkgPath()
	typeString := t.TypeString()

	local = t.GetLocalTypeName()
	prefix = strings.TrimSuffix(typeString, fullPath+"."+local)
	path = t.PkgPath()

	return
}

func traverseType(typ types.Type) (path string) {
	switch t := typ.(type) {
	case *types.Basic:
		return
	case *types.Named:
		if t.Obj().Pkg() == nil {
			return
		}
		return t.Obj().Pkg().Path()

	case *types.Array:
		return traverseType(t.Elem())
	case *types.Slice:
		return traverseType(t.Elem())
	case *types.Pointer:
		return traverseType(t.Elem())
	case *types.Map:
		return traverseType(t.Elem())

	case *types.Struct:
	case *types.Tuple:
	case *types.Signature:
	case *types.Interface:
	case *types.Chan:
	}
	log.Println("WARN Using unsuported type", reflect.TypeOf(typ))
	return
}

func (fn *Function) GetTypeDeps() (deps []Field) {
	//fields := append(fn.Arguments, fn.Results...)
	//for _, field := range fields {
	//}
	return
}
