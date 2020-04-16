package parser

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
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
	Pkg     *types.Package
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
	log.Println(">", pkg)
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

	var files []*ast.File
	for _, file := range p.pkg.Files {
		files = append(files, file)
	}

	conf := types.Config{Importer: importer.For("source", nil)}

	p.Pkg, err = conf.Check(p.Package.Path, p.fset, files, nil)
	if err != nil {
		log.Println("ERR parsing", err)
	}

	return nil
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

//GetFunctions returns exported functions from package
func (p *Parser) GetFunctions() (functions []*Function) {
	scope := p.Pkg.Scope()
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		switch f := o.(type) {
		case *types.Func:
			if !f.Exported() {
				continue
			}
			sig := f.Type().(*types.Signature)
			args := p.extractArgsList(sig.Params())
			results := p.extractArgsList(sig.Results())
			var receiver Field
			for _, rec := range p.extractArgsList(types.NewTuple(sig.Recv())) {
				receiver = rec
			}

			function := &Function{
				Name:        f.Name(),
				Arguments:   args,
				Results:     results,
				Receiver:    receiver,
				Package:     p.Service.Alias,
				ServiceType: p.Service.Type,
			}
			functions = append(functions, function)
		}
	}
	return
}

func (p *Parser) extractArgsList(list *types.Tuple) (args []Field) {
	length := list.Len()
	if list == nil || length == 0 {
		return
	}

	for count := 0; count < length; count++ {
		v := list.At(count)
		if v == nil {
			continue
		}

		name := v.Name()
		if name == "" {
			name = fmt.Sprintf("arg%d", count)
		}

		field := Field{
			Name: name,
			Var:  v,
			Type: Type{v.Type()},
		}
		args = append(args, field)
	}

	return
}

//GetTypes returns exported sturct types from package
func (p *Parser) GetTypes() (specs []*TypeSpec, err error) {
	scope := p.Pkg.Scope()
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		switch t := o.(type) {
		case *types.TypeName:
			log.Println(">>>>>TYPE", t.Type().String())
		}
	}
	return
}

