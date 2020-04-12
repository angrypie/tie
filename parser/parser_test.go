package parser

import (
	"testing"

	"github.com/angrypie/tie/types"
)

func TestParser(t *testing.T) {
	pkg := "github.com/angrypie/tie/example/receiver"
	service := &types.Service{Name: pkg}
	parser := NewParser(service)
	err := parser.Parse(pkg)
	if err != nil {
		t.Error(err)
	}

	functions := parser.GetFunctions()
	if len(functions) == 0 {
		t.Error("GetFunctions should return more than 0 functions")
	}
}
