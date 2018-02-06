package template

import (
	"bytes"
	"text/template"

	"github.com/angrypie/tie/parser"
)

const ServerMain = `
func main() {

	err := {{.Alias}}.InitService()
	if err != nil {
		fmt.Println("Cant InitService", err)
		return
	}

	{{if ne .ServiceType "httpOnly"}}
	go startRPCServer()
	{{end}}

	{{if eq .ServiceType "http"}}
	go startHTTPServer()
	{{end}}

	//TODO graceful shutdown
	<-make(chan bool)
}

{{if eq .ServiceType "http"}}
func startHTTPServer() {
	port, err := getPort()
	if err != nil {
		panic(err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	e := echo.New()
	{{range $k,$v := .Functions}}e.POST(strings.ToLower("{{$v.Name}}"), {{$v.Name}}HTTPHandler)
	{{end}}
	e.Start(addr)
}
{{end}}

{{if ne .ServiceType "httpOnly"}}
func startRPCServer() {
	port, err := getPort()
	if err != nil {
		panic(err)
	}

	fmt.Println("Resource_{{.Alias}}")
	zconfServer, err := zeroconf.Register("GoZeroconf", "Resource_{{.Alias}}", "local.", port, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		panic(err)
	}
	defer zconfServer.Shutdown()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := rpcx.NewServer()
	server.RegisterName("Resource_{{.Alias}}", new(Resource_{{.Alias}}))
	fmt.Println("Start on port:", port)
	err = server.Serve("tcp", addr)
	if err != nil {
		panic(err)
	}
}
{{end}}

func getPort() (port int, err error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return port, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return port, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
`

func MakeServerMain(p *parser.Parser, functions []*parser.Function) ([]byte, error) {
	type helper struct {
		Alias       string
		ServiceType string
		Functions   []*parser.Function
	}
	h := helper{Alias: p.Package.Alias, ServiceType: p.ServiceType, Functions: functions}

	var buff bytes.Buffer
	t := template.Must(
		template.New("server_main").Parse(ServerMain),
	)
	err := t.Execute(&buff, h)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
