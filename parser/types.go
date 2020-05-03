package parser

import (
	"go/build"
	"go/types"
	"log"
	"reflect"
	"strings"
)

type ResultFields struct {
	Last Field
	body []Field
}

func (rf ResultFields) List() []Field {
	return append(rf.body, rf.Last)
}

type Function struct {
	Name        string
	Arguments   []Field
	Results     ResultFields
	Receiver    Field
	Package     string
	ServiceType string
}

type TypeSpec struct {
	Name   string
	Fields []Field
}

type StructType struct {
	Name   string
	Fields []Field
}

func NewStructType(name string, fields []Field) StructType {
	return StructType{
		Name:   name,
		Fields: fields,
	}
}

type Field struct {
	name string
	Var  *types.Var
	Type
}

//IsDefined return true if Var type is set.
func (field Field) IsDefined() bool {
	return field.Var != nil
}

func (field Field) Name() string {
	return field.name
}

type Type struct {
	typ types.Type
}

func (t Type) TypeName() string {
	arr := strings.Split(t.typ.String(), t.fullPkgPath()+".")
	return arr[len(arr)-1]
}

func (t Type) TypeParts() (prefix, path, local string) {
	fullPath := t.fullPkgPath()
	typeString := t.typ.String()

	local = t.TypeName()
	prefix = strings.TrimSuffix(typeString, fullPath+"."+local)
	path = strings.TrimPrefix(
		t.fullPkgPath(),
		build.Default.GOPATH+"/src/",
	)

	return
}

func (t Type) fullPkgPath() string {
	return traverseType(t.typ)
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
		return
	case *types.Interface:
	case *types.Chan:
	}
	log.Println("WARN Using unsuported type", reflect.TypeOf(typ), typ.String())
	return
}
