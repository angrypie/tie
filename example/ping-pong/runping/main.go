package main

import (
	"log"

	"github.com/angrypie/tie/example/ping-pong/ping"
)

func main() {
	ret, err := ping.Ping()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(ret)
}
