//Server api for package: github.com/angrypie/tie/example/basic
//absolute path: /home/el/go/src/github.com/angrypie/tie/example/basic
//package alias: basic

package main

import (
	//import original package
	"github.com/angrypie/tie/example/basic"

	//import RPCX package
	"github.com/smallnest/rpcx"
)

//Main api resource (for pure functions)
type Resource_basic struct{}

type MulRequest struct {
	A int
	B int
}

type MulResponse struct {
	Result int
	Err    error
}

func (r *Resource_basic) Mul(request *MulRequest, response *MulResponse) (err error) {
	//1. Call original function

	Result, Err := basic.Mul(
		request.A,
		request.B,
	)
	//2. Put results to response struct
	response.Result = Result
	response.Err = Err

	//3. Return error or nil
	return err
}

func main() {
	addr := "127.0.0.1:9999"
	server := rpcx.NewServer()
	server.RegisterName("Resource_basic", new(Resource_basic))
	server.Serve("tcp", addr)
}
