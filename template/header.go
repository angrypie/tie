package template

import (
	"bytes"
	"html/template"

	"github.com/angrypie/tie/parser"
)

const ServerHeader = `
//Server api for package: {{.Package.Name}}
//absolute path: {{.Package.Path}}
//package alias: {{.Package.Alias}}
package main
import (
	//import original package
	"{{.Package.Name}}"
)
`

func MakeServerHeader(p *parser.Parser) ([]byte, error) {
	var buff bytes.Buffer
	t := template.Must(
		template.New("server_header").Parse(ServerHeader),
	)
	err := t.Execute(&buff, p)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
