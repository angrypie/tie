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

func NewParser(service *tieTypes.Service) *Parser {
	fset := token.NewFileSet()
	return &Parser{
		fset:    fset,
		Service: service,
	}
}

func (p *Parser) Parse(pkg string) error {
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
		return errors.New("Parsed directory should contain one package")
	}

	p.pkgs = pkgs
	for _, pkg := range pkgs {
		p.pkg = pkg
		break
	}
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

//TODO Refactoring with GetFunctions
func (p *Parser) GetTypes() (types []*Type, err error) {
	for _, pkg := range p.pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.GenDecl:
					if n.Tok != token.TYPE {
						break
					}
					//TODO process every spec
					ts := n.Specs[0].(*ast.TypeSpec)

					if st, ok := ts.Type.(*ast.StructType); ok {
						if t, ok := p.processType(st, ts); ok {
							log.Println(1, ts.Name.Name)
							types = append(types, t)
						}
					}

					if st, ok := ts.Type.(*ast.SelectorExpr); ok {
						log.Println(2, st.Sel.String())
						fmt.Printf("%+v\n", st)
						fmt.Printf("%+v\n", st.Sel)
						fmt.Printf("%+v\n", st.X)

					}

				}
				return true
			})
		}
	}
	return types, nil
}

//ToFiles returns array of files in package. Each file represents as a bytes array.
func (p *Parser) ToFiles() (files [][]byte) {
	for _, pkg := range p.pkgs {
		for _, file := range pkg.Files {
			var buf bytes.Buffer
			printer.Fprint(&buf, p.fset, file)
			files = append(files, buf.Bytes())
		}
	}
	return files
}

//UpgradeApiImports returns false if import deleted but not added.
func (p *Parser) UpgradeApiImports(imports []string) bool {

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

func (p *Parser) GetPackageName() string {
	return p.pkg.Name
}
