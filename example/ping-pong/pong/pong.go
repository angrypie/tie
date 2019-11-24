package pong

import "github.com/angrypie/tie/example/ping-pong/health"

func Pong() (string, error) {
	health.Check("Pong")
	return "pong", nil
}
