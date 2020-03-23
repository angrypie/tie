package parser

import (
	"testing"

	"github.com/angrypie/tie/types"
	"github.com/stretchr/testify/assert"
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

func TestHelpers(t *testing.T) {
	assert := assert.New(t)
	prefixes := []string{
		"*", "***", "[]", "[][][][]", "*[]", "[]*", "*[][][]*[]", "",
	}
	for _, expectedPrefix := range prefixes {
		prefix := getTypePrefix(expectedPrefix + "T")
		assert.Equal(expectedPrefix, prefix)
	}
}
