package template

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/angrypie/tie/parser"
)

const ServerHeader = `
package api
`

const ApiWrapper = `
func {{.Name}}(
	{{range $k,$v := .Arguments}}
	{{$v.Name}} {{$v.Type}},
	{{end}}
) (
	{{range $k,$v := .Results}}
	{{$v.Name}} {{$v.Type}},
	{{end}}
){

}
`

func MakeApiWrapper(fn *parser.Function) ([]byte, error) {
	if fn == nil {
		return nil, errors.New("fn must be not nil")
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
