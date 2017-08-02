package tasks

import "testing"

func TestConfig(t *testing.T) {
	config := []byte(`
services:
  - name: 'github.com/angrypie/tie/example/basic'
    alias: 'arith'
  - name: 'github.com/angrypie/tie/example/basic/usage'
    alias: '7mul6'`)

	err := Config(config)
	if err != nil {
		t.Error(err)
	}

}
