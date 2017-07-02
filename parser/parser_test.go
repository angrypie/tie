package parser

import "testing"

func TestParser(t *testing.T) {
	parser := NewParser()
	err := parser.Parse("github.com/angrypie/tie/parser")
	if err != nil {
		t.Error(err)
	}

	functions, err := parser.GetFunctions()
	if err != nil {
		t.Error(err)
	}

	if len(functions) == 0 {
		t.Error("GetFunctions should return more than 0 functions")
	}
}
