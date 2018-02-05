package template

import (
	"bytes"
	"text/template"

	"github.com/angrypie/tie/parser"
)

const ServerMain = `
func main() {
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

	e := echo.New()
	for path, fn := range echoEndpoints {
		echo.POST(path, fn)
	}

}

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
