package parser

import (
	"log"
	"testing"
)

func TestParser(t *testing.T) {
	parser := NewParser()
	err := parser.Parse(".")
	if err != nil {
		t.Error(err)
	}
	functions, err := parser.GetFunctions()
	if err != nil {
		t.Error(err)
	}
	log.Println(functions)

}
