package pong

import (
	"log"

	"github.com/angrypie/tie/example/ping-pong/health"
)

func Pong() (resp string, err error) {
	err = health.Check("Pong")
	if err != nil {
		log.Println("ERR health check", err)
	}
	return "pong", nil
}
