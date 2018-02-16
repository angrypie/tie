package template

import (
	"bytes"
	"text/template"

	"github.com/angrypie/tie/parser"
)

const ServerMain = `
func main() {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		<-gracefulStop
		{{if .IsStopService}}
		err := {{.Alias}}.StopService()
		if err != nil {
			fmt.Println("ERR Cant gracefully stop service", err)
		}
		{{end}}
		os.Exit(0)
	}()

	{{if .IsInitService}}
	err := {{.Alias}}.InitService()
	if err != nil {
		fmt.Println("ERR Cant InitService", err)
		return
	}
	{{end}}

	{{if ne .ServiceType "httpOnly"}}
	go startRPCServer()
	{{end}}

	{{if or (eq .ServiceType "http") (eq .ServiceType "httpOnly")}}
	go startHTTPServer()
	{{end}}

	//TODO graceful shutdown
	<-make(chan bool)
}

{{if or (eq .ServiceType "http") (eq .ServiceType "httpOnly")}}
func startHTTPServer() {
	//port, err := getPort()
	//if err != nil {
		//panic(err)
	//}
	addr := fmt.Sprintf(":%d", 8111)
	e := echo.New()
	{{range $k,$v := .Functions}}e.POST(strings.ToLower("{{$v.Name}}"), {{$v.Name}}HTTPHandler)
	{{end}}
	e.Start(addr)
}
{{end}}

{{if or (eq .ServiceType "http") (eq .ServiceType "httpOnly")}}
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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
		Alias         string
		ServiceType   string
		Functions     []*parser.Function
		IsInitService bool
		IsStopService bool
	}
	var fns []*parser.Function
	for _, fn := range functions {
		if name := fn.Name; name == "InitService" || name == "StopService" {
			continue
		}
		fns = append(fns, fn)
	}
	h := helper{Alias: p.Package.Alias, ServiceType: p.ServiceType, Functions: fns}

	for _, fn := range functions {
		if fn.Name == "InitService" {
			h.IsInitService = true
		}
		if fn.Name == "StopService" {
			h.IsStopService = true
		}
	}

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
