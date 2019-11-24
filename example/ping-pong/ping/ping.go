package ping

import (
	"github.com/angrypie/tie/example/ping-pong/health"
	"github.com/angrypie/tie/example/ping-pong/pong"
)

func Ping() (ret string, err error) {
	health.Check("Ping")
	ret, err = pong.Pong()
	return "ping-" + ret, err
}
