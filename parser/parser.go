package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
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

//DEPRECATED
func (p *Parser) ReplaceImport(from, to string) (ok bool, files []bytes.Buffer) {
	ok = true
	arr := strings.Split(from, "/")
	alias := arr[len(arr)-1]

	for _, pkg := range p.pkgs {
		for _, file := range pkg.Files {
			ok := astutil.DeleteImport(p.fset, file, from)
			if ok {
				ok = astutil.AddNamedImport(p.fset, file, alias, to)
				if !ok {
					return false, nil
				}
			}
			var buf bytes.Buffer
			printer.Fprint(&buf, p.fset, file)
			files = append(files, buf)
		}
	}
	return true, files
}

func (p *Parser) UpgradeApiImports(imports []string) (ok bool) {
	ok = true

	for _, pkg := range p.pkgs {
		for _, file := range pkg.Files {
			for _, path := range imports {
				//get alias from path
				//TODO support named ipmports
				arr := strings.Split(path, "/")
				alias := arr[len(arr)-1]
				ok := astutil.DeleteImport(p.fset, file, path)
				if ok {
					ok = astutil.AddNamedImport(p.fset, file, alias, path+"/tie_client")
					if !ok {
						return false
					}
				}
			}
		}
	}
	return true
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
