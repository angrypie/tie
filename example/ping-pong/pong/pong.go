package pong

import "github.com/angrypie/tie/example/ping-pong/health"

func Pong() (resp string, err error) {
	health.Check("Pong")
	return "pong", nil
}
