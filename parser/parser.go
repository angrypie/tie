package parser

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type Package struct {
	Name  string
	Alias string
	Path  string
}

type Parser struct {
	fset    *token.FileSet
	pkgs    map[string]*ast.Package
	Package *Package
}

func NewParser() *Parser {
	fset := token.NewFileSet()
	return &Parser{
		fset: fset,
	}
}

func (p *Parser) Parse(pkg string) error {
	p.Package = NewPackage(pkg)
	pkgs, err := parser.ParseDir(
		p.fset, p.Package.Path, func(info os.FileInfo) bool {
			name := info.Name()
			return !info.IsDir() &&
				!strings.HasPrefix(name, ".") &&
				strings.HasSuffix(name, ".go") &&
				!strings.HasSuffix(name, "_test.go")
		}, parser.ParseComments)
	if err != nil {
		return err
	}
	p.pkgs = pkgs
	return nil
}

func (p *Parser) GetFunctions() (functions []*Function, err error) {
	for _, pkg := range p.pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.FuncDecl:
					if function, ok := p.processFunction(n); ok {
						functions = append(functions, function)
					}
				}
				return true
			})
		}
	}
	return functions, nil
}

func NewPackage(name string) *Package {
	arr := strings.Split(name, "/")
	alias := arr[len(arr)-1]
	return &Package{
		Name:  name,
		Alias: alias,
		Path:  fmt.Sprintf("%s/src/%s", build.Default.GOPATH, name),
	}
}
