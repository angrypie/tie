package main

import (
	"fmt"

	"github.com/angrypie/tie/example/ping-pong/ping"
)

func main() {
	ret, err := ping.Ping()
	fmt.Println(ret, err)
}
