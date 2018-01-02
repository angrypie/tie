package template

import (
	"bytes"
	"log"
	"text/template"

	"github.com/angrypie/tie/parser"
)

const ExportedType = `
type {{.Name}} struct {
	{{range $k,$v := .Fields}}{{$v.Name}} {{$v.Type}}
	{{end}}
}
`

//TODO STUB
func MakeClientTypes(p *parser.Parser) ([]byte, error) {
	var buff bytes.Buffer
	types, err := p.GetTypes()

	for _, exportedType := range types {
		log.Println(exportedType.Name)
		if err != nil {
			return nil, err
		}
		t := template.Must(
			template.New("exported_type").Parse(ExportedType),
		)
		//TODO it shoud apend data to buff, need check
		err := t.Execute(&buff, exportedType)
		if err != nil {
			return nil, err
		}
	}

	return buff.Bytes(), nil
}
