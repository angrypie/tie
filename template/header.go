package template

import (
	"bytes"
	"text/template"

	"github.com/angrypie/tie/parser"
)

const ServerHeader = `
//Server api for package: {{.Package.Name}}
//absolute path: {{.Package.Path}}
//package alias: {{.Package.Alias}}

package main
import (
	//import original package
	{{.Package.Alias}} "{{.Package.Name}}/tie_upgraded"

	{{if ne .ServiceType "httpOnly"}}
	//import RPCX package
	"github.com/smallnest/rpcx"
	"github.com/grandcat/zeroconf"
	{{end}}

	//import util packages
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	//import http packages
{{if or (eq .ServiceType "http") (eq .ServiceType "httpOnly")}}
	"github.com/labstack/echo"
	"strings"
	"net/http"
	{{end}}
)

{{if ne .ServiceType "httpOnly"}}
//Main api resource (for pure functions)
type Resource_{{.Package.Alias}} struct {}
{{end}}
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

const ClientHeader = `
//Client api for package: {{.Package.Name}}
//absolute path: {{.Package.Path}}
//package alias: {{.Package.Alias}}

package {{.Package.Alias}}_api
import (
	//import RPCX package
	"github.com/smallnest/rpcx"
	"context"
	"time"
	"fmt"
	"github.com/grandcat/zeroconf"
	"errors"
)
func getLocalService(service string) (port int, err error) {

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return port, err
	}

	entries := make(chan *zeroconf.ServiceEntry)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = resolver.Browse(ctx, service, "local.", entries)
	if err != nil {
		return port, err
	}

	select {
	case <-ctx.Done():
		return port, errors.New("Service not found")

	case entry := <-entries:
		return entry.Port, nil
	}
}
`

func MakeClientHeader(p *parser.Parser) ([]byte, error) {
	var buff bytes.Buffer
	t := template.Must(
		template.New("client_header").Parse(ClientHeader),
	)
	err := t.Execute(&buff, p)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
