package parser

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	tieTypes "github.com/angrypie/tie/types"
	"github.com/spf13/afero"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/types/typeutil"
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
func (p *Parser) Parse(pkgPath string) error {
	log.Println(">", pkgPath)

	modPath, err := GetModulePath(pkgPath)
	if err != nil {
		return err
	}

	p.Package = NewPackage(pkgPath, modPath)
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

	conf := types.Config{Importer: importer.ForCompiler(p.fset, "source", nil)}

	p.Pkg, err = conf.Check(p.Package.Path, p.fset, files, nil)
	if err != nil {
		log.Println("ERR parsing", err)
	}

	return nil
}

type File struct {
	Name    string
	Content []byte
}

//ToFiles returns array of files in package. Each file represents as a bytes array.
func (p *Parser) ToFiles() (files []File) {
	for path, file := range p.pkg.Files {
		var buf bytes.Buffer
		printer.Fprint(&buf, p.fset, file)
		_, name := filepath.Split(path)
		files = append(files, File{
			Name:    name,
			Content: buf.Bytes(),
		})
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
func NewPackage(path string, modulePath string) *Package {
	arr := strings.Split(path, "/")
	alias := arr[len(arr)-1]
	return &Package{
		Name:  modulePath,
		Alias: alias,
		Path:  path,
	}
}

//GetPackageName returns package name.
func (p *Parser) GetPackageName() string {
	return p.pkg.Name
}

//GetFunctions returns exported functions from package
func (p *Parser) GetFunctions() (functions []Function) {
	addFunc := func(f *types.Func) {
		if !f.Exported() {
			return
		}
		sig := f.Type().(*types.Signature)
		receiver := NewField(sig.Recv())
		args := extractArgsList(sig.Params())
		results, err := resultsFromArgs(extractArgsList(sig.Results()))
		if err != nil {
			log.Printf("skip function %s: %e\n", f.FullName(), err)
			return
		}

		function := Function{
			Name:        f.Name(),
			Arguments:   args,
			Results:     results,
			Receiver:    receiver,
			Package:     p.Service.Alias,
			ServiceType: p.Service.Type,
		}
		functions = append(functions, function)
	}
	scope := p.Pkg.Scope()
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		switch t := o.(type) {
		case *types.TypeName:
			mset := &typeutil.MethodSetCache{}
			methods := typeutil.IntuitiveMethodSet(t.Type(), mset)
			for _, method := range methods {
				addFunc(method.Obj().(*types.Func))
			}
		case *types.Func:
			addFunc(t)
		}
	}
	return
}

//resultsFromArgs creates field list that contain error field type at last position
func resultsFromArgs(args []Field) (results ResultFields, err error) {
	length := len(args)
	if length == 0 {
		return
	}
	last := args[length-1]
	if last.TypeName() != "error" {
		err = errors.New("method should have (err error) return type at last position")
		return
	}

	results = ResultFields{Last: last, body: args[0 : length-1]}
	return
}

func extractArgsList(list *types.Tuple) (args []Field) {
	if list == nil {
		return
	}

	for count, length := 0, list.Len(); count < length; count++ {
		v := list.At(count)
		if v == nil {
			continue
		}

		field := NewField(v)

		if field.name == "" {
			field.name = fmt.Sprintf("arg%d", count)
		}

		args = append(args, field)
	}

	return
}

func NewField(v *types.Var) Field {
	if v == nil {
		return Field{}
	}
	return Field{
		name: v.Name(),
		Var:  v,
		Type: Type{v.Type()},
	}
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

func GetModulePath(dirPath string) (modPath string, err error) {
	fs := afero.NewOsFs()
	gomod := path.Join(dirPath, "go.mod")
	buf, err := afero.ReadFile(fs, gomod)
	if err != nil {
		return modPath, fmt.Errorf("reading go.mod %s: %w", gomod, err)
	}

	modPath = modfile.ModulePath(buf)
	if modPath == "" {
		return modPath, fmt.Errorf("canont find a module path for %s", gomod)
	}

	return
}
