package pong

import "github.com/angrypie/tie/example/ping-pong/helth"

func Pong() (string, error) {
	helth.Check("Pong")
	return "pong", nil
}
