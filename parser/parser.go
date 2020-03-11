package parser

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"

	tieTypes "github.com/angrypie/tie/types"
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
	pkg     *ast.Package
	Package *Package
	Service *tieTypes.Service
}

//NewParser creates new parser.
func NewParser(service *tieTypes.Service) *Parser {
	fset := token.NewFileSet()
	return &Parser{
		fset:    fset,
		Service: service,
	}
}

//Parse initializes parser by parsing package. Should be called before any other method.
func (p *Parser) Parse(pkg string) error {
	log.Println(pkg)
	p.Package = NewPackage(pkg)
	if p.Service.Alias == "" {
		p.Service.Alias = p.Package.Alias
	}
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

	if len(pkgs) != 1 {
		return errors.New("parsed directory should contain one package (TODO)")
	}

	p.pkgs = pkgs
	for _, pkg := range pkgs {
		p.pkg = pkg
		break
	}

	return nil
}

func inspectNodesInPkg(pkg *ast.Package, inspect func(node ast.Node) bool) {
	for _, file := range pkg.Files {
		ast.Inspect(file, inspect)
	}
}

//GetFunctions returns exported functions from package
func (p *Parser) GetFunctions() (functions []*Function) {
	inspectNodesInPkg(p.pkg, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			if function, ok := p.processFunction(n); ok {
				functions = append(functions, function)
			}
		}
		return true
	})
	return
}

//GetTypes returns exported types from package
func (p *Parser) GetTypes() (types []*Type, err error) {
	inspectNodesInPkg(p.pkg, func(node ast.Node) (goInDepth bool) {
		switch n := node.(type) {
		case *ast.GenDecl:
			if n.Tok != token.TYPE {
				return
			}
			for _, spec := range n.Specs {
				ts := spec.(*ast.TypeSpec)
				//TODO: handle other type specs
				if st, ok := ts.Type.(*ast.StructType); ok {
					if t, ok := p.processType(st, ts); ok {
						types = append(types, t)
					}
				}
			}
		}
		return
	})
	return types, nil
}

//ToFiles returns array of files in package. Each file represents as a bytes array.
func (p *Parser) ToFiles() (files [][]byte) {
	for _, file := range p.pkg.Files {
		var buf bytes.Buffer
		printer.Fprint(&buf, p.fset, file)
		files = append(files, buf.Bytes())
	}
	return files
}

//UpgradeApiImports returns false if import deleted but not added.
func (p *Parser) UpgradeApiImports(imports []string, upgrade func(string) string) bool {
	allImports := make(map[string]struct{})

	for _, path := range imports {
		allImports[path] = struct{}{}
	}

	for _, file := range p.pkg.Files {
		for _, par := range astutil.Imports(p.fset, file) {
			for _, i := range par {
				path := strings.Trim(i.Path.Value, `"`)
				if _, ok := allImports[path]; !ok {
					continue
				}

				var alias, name string
				if i.Name == nil {
					//get import name  from path
					arr := strings.Split(path, "/")
					alias = arr[len(arr)-1]
				} else {
					name = i.Name.Name
					alias = name
				}

				if astutil.DeleteNamedImport(p.fset, file, name, path) {
					if !astutil.AddNamedImport(p.fset, file, alias, upgrade(path)) {
						return false
					}
				}
			}
		}
	}

	return true
}

//NewPackage returns new Packege instance wit initialized Name, Alias and Path.
func NewPackage(name string) *Package {
	arr := strings.Split(name, "/")
	alias := arr[len(arr)-1]
	return &Package{
		Name:  name,
		Alias: alias,
		Path:  fmt.Sprintf("%s/src/%s", build.Default.GOPATH, name),
	}
}

//GetPackageName returns package name.
func (p *Parser) GetPackageName() string {
	return p.pkg.Name
}
