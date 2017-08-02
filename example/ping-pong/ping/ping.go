package ping

import (
	"github.com/angrypie/tie/example/ping-pong/helth"
	"github.com/angrypie/tie/example/ping-pong/pong"
)

func Ping() (ret string, err error) {
	helth.Check("Ping")
	if ret, err := pong.Pong(); err != nil {
		return "ping-" + ret, nil
	} else {
		return "ping-" + ret, nil
	}
}
