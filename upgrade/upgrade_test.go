package upgrade

import (
	"io/ioutil"
	"testing"
)

func TestUpgrade(t *testing.T) {
	upgrade, err := Server("github.com/angrypie/tie/example/basic")
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile("/tmp/dat1", upgrade.Server.Bytes(), 0644)
	if err != nil {
		t.Error(err)
	}
}
