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

type Type struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Var  *types.Var
}

func (field Field) TypeString() string {
	return field.Var.Type().String()
}

func (field Field) GetLocalTypeName() string {
	arr := strings.Split(field.Var.Type().String(), field.Var.Pkg().Path()+".")
	return arr[len(arr)-1]
}

//IsDefined return true if Var type is set.
func (field Field) IsDefined() bool {
	return field.Var != nil
}

func (field Field) PkgPath() string {
	return strings.TrimPrefix(
		field.fullPkgPath(),
		build.Default.GOPATH+"/src/",
	)
}

func (field Field) fullPkgPath() string {
	return traverseType(field.Var.Type())
}

func (field Field) GetTypeParts() (prefix, path, local string) {
	fullPath := field.fullPkgPath()
	arr := strings.Split(field.TypeString(), fullPath+".")

	local = arr[len(arr)-1]
	prefix = strings.TrimSuffix(field.TypeString(), fullPath+"."+local)
	path = field.PkgPath()

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
