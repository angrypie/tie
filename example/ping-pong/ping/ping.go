package ping

import (
	"log"

	"github.com/angrypie/tie/example/ping-pong/health"
	"github.com/angrypie/tie/example/ping-pong/pong"
)

func Ping() (ret string, err error) {
	err = health.Check("Ping")
	if err != nil {
		log.Println("ERR health check", err)
	}
	ret, err = pong.Pong()
	return "ping-" + ret, err
}
