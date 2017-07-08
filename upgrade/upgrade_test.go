package upgrade

import (
	"os"
	"testing"
)

func TestUpgrade(t *testing.T) {
	upgrade, err := Server("github.com/angrypie/tie/example/basic")
	if err != nil {
		t.Error(err)
	}

	_, err = upgrade.Server.WriteTo(os.Stdout)
	if err != nil {
		t.Error(err)
	}
}
