package template

import (
	"bytes"
	"errors"
	"html/template"
	"strings"

	"github.com/angrypie/tie/parser"
)

const ApiWrapper = `
type {{.Name}}Request struct {
	{{range $k,$v := .Arguments}}{{$v.Name}} {{$v.Type}}
	{{end}}
}

type {{.Name}}Response struct {
	{{range $k,$v := .Results}}{{$v.Name}} {{$v.Type}}
	{{end}}
}

func (r *Resource_{{.Package}}) {{.Name}}(request *{{.Name}}Request, response *{{.Name}}Response) (err error) {
	//1. Call original function


	{{range $k,$v := .Results}}{{if $k}},{{end}} {{$v.Name}}{{end}} := {{.Package}}.{{.Name}}(
		{{range $k,$v := .Arguments}}request.{{$v.Name}},
		{{end}}
	)
	//2. Put results to response struct
	{{range $k,$v := .Results}}response.{{$v.Name}} = {{$v.Name}}
	{{end}}
	//3. Return error or nil
	return err
}
`

func MakeApiWrapper(fn *parser.Function) ([]byte, error) {
	if fn == nil {
		return nil, errors.New("fn must be not nil")
	}

	for i, _ := range fn.Arguments {
		fn.Arguments[i].Name = strings.Title(fn.Arguments[i].Name)
	}

	for i, _ := range fn.Results {
		fn.Results[i].Name = strings.Title(fn.Results[i].Name)
	}

	var buff bytes.Buffer
	t := template.Must(
		template.New("api_wrapper").Parse(ApiWrapper),
	)
	err := t.Execute(&buff, fn)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
