package template

import (
	"bytes"
	"html/template"

	"github.com/angrypie/tie/parser"
)

const ServerMain = `
func main() {
	addr := "127.0.0.1:9999"
	server := rpcx.NewServer()
	server.RegisterName("Resource_{{.Alias}}", new(Resource_{{.Alias}}))
	server.Serve("tcp", addr)
}
`

func MakeServerMain(p *parser.Package) ([]byte, error) {
	var buff bytes.Buffer
	t := template.Must(
		template.New("server_main").Parse(ServerMain),
	)
	err := t.Execute(&buff, p)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
