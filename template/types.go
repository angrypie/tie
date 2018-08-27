package template

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/angrypie/tie/parser"
)

//TODO Listing arguments is almost same for api functions,
//api types, and exported types, need to generalize this logic
const ExportedType = `
type {{.Name}} struct {
	{{range $k,$v := .Fields}}{{$v.Name}} {{$v.Prefix}}{{if $v.Package}}{{$v.Package}}.{{end}}{{$v.Type}} {{$v.Name | tolower | printf "json:%q" | tobackquote}}
	{{end}}
}
`

func MakeClientTypes(p *parser.Parser) ([]byte, error) {
	var buff bytes.Buffer
	types, err := p.GetTypes()
	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"tolower": strings.ToLower,
		"tobackquote": func(str string) string {
			return "`" + str + "`"
		},
	}

	for _, exportedType := range types {
		if err != nil {
			return nil, err
		}
		t := template.Must(
			template.New("exported_type").Funcs(funcMap).Parse(ExportedType),
		)

		err := t.Execute(&buff, exportedType)
		if err != nil {
			return nil, err
		}
	}

	return buff.Bytes(), nil
}
