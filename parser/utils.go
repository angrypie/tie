package parser

import (
	"go/ast"
	"regexp"
)

func (p *Parser) processType(st *ast.StructType, ts *ast.TypeSpec) (*Type, bool) {
	t := &Type{
		Name: ts.Name.Name,
		//Fields: p.extractArgsList(st.Fields),
	}

	return t, true
}

func isExportedType(t string) bool {
	return ast.IsExported(t)
}

func getTypePrefix(t string) (prefix string) {
	re := regexp.MustCompile(`[^\[\]\*].*$`)
	split := re.Split(t, -1)
	return split[0]
}
