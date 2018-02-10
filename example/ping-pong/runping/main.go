package main

import (
	"fmt"

	"github.com/angrypie/tie/example/ping-pong/ping"
)

func main() {
	ret, err := ping.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)
}
