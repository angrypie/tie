package template

import (
	"bytes"
	"errors"
	"strings"
	"text/template"

	"github.com/angrypie/tie/parser"
)

const ApiClient = `
type {{.Name}}Request struct {
	{{range $k,$v := .Arguments}}{{$v.Name}} {{$v.Type}}
	{{end}}
}

type {{.Name}}Response struct {
	{{range $k,$v := .Results}}{{$v.Name}} {{$v.Type}}
	{{end}}
}

func {{.Name}}(
	{{range $k,$v := .Arguments}}{{$v.Name}} {{$v.Type}},
	{{end}}
) (
	{{range $k,$v := .Results}}{{$v.Name}} {{$v.Type}},
	{{end}}
) {
	port, Err := getLocalService("Resource_{{.Package}}")
	if Err != nil {
		return {{range $k,$v := .Results}}{{if $k}},{{end}} {{$v.Name}}{{end}}
	}
	s := &rpcx.DirectClientSelector{
		Network: "tcp",
		Address: fmt.Sprintf("127.0.0.1:%d", port),
		DialTimeout: 2 * time.Second,
	}
	client := rpcx.NewClient(s)

	request := &{{.Name}}Request{
		{{range $k,$v := .Arguments}}{{$v.Name}},
		{{end}}
	}

	var response {{.Name}}Response

	client.Call(context.Background(), "Resource_{{.Package}}.{{.Name}}", request, &response)
	client.Close()
	return {{range $k,$v := .Results}}{{if $k}},{{end}} response.{{$v.Name}}{{end}}
}
`
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

//Http handlers
{{if eq .ServiceType "http"}}

func  {{.Name}}HTTPHandler(c echo.Context) (err error) {
	//1. Bind request params
	{{if .Arguments}}
	request := new({{.Name}}Request)

	if err := c.Bind(request); err != nil {
		return err
	}

	fmt.Println("Request", request)
	{{end}}


	//2. Call original function
	{{range $k,$v := .Results}}{{if $k}},{{end}} {{$v.Name}}{{end}} := {{.Package}}.{{.Name}}(
		{{range $k,$v := .Arguments}}request.{{$v.Name}},
		{{end}}
	)

	response := new({{.Name}}Response)
	//3. Put results to response struct
	{{range $k,$v := .Results}}response.{{$v.Name}} = {{$v.Name}}
	{{end}}

	fmt.Println("Response", response)

	return c.JSON(http.StatusOK, response)
}

{{end}}
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

//TODO refactor
func MakeApiClient(fn *parser.Function) ([]byte, error) {
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
		template.New("api_client").Parse(ApiClient),
	)
	err := t.Execute(&buff, fn)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
