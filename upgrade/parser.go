package upgrade

import "go/ast"

type Visitor struct {
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	return v.typeSelector(n)
}

func (v *Visitor) typeSelector(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.FuncDecl:
		CreateApi(n.Name.Name)
	default:
		return v
	}

	return nil
}
