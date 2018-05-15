package tasks

import "testing"

func TestConfig(t *testing.T) {
	config := []byte(`
services:
  - name: 'github.com/angrypie/tie/example/ping-pong/ping'
    alias: 'ping'
  - name: 'github.com/angrypie/tie/example/ping-pong/pong'
    alias: 'pong'
  - name: 'github.com/angrypie/tie/example/ping-pong/helth'
    alias: 'helth'
  - name: 'github.com/angrypie/tie/example/ping-pong/runping'
    alias: 'runping'
`)

	err := ConfigFromYaml(config, ".")
	if err != nil {
		t.Error(err)
	}
}

func TestReadDir(t *testing.T) {
	err := ReadDirAsConfig("../example/basic")
	if err != nil {
		t.Error(err)
	}

	err = CleanBinary("../example/basic")
	if err != nil {
		t.Error(err)
	}
}
