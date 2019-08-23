package receiver

import (
	"testing"
)

func TestReceiver(t *testing.T) {
	user := User{}
	user.InitReceiver(func(header string) string {
		return "Paul"
	})

	greeting, err := user.Greeting()
	if err != nil {
		t.Error(err)
	}
	if greeting != "Hello, my name is Paul" {
		t.Error("Greetings does not match" + greeting)
	}
}
