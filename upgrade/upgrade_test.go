package upgrade

import "testing"

func TestUpgrade(t *testing.T) {
	upgrade, err := Server("github.com/angrypie/tie/example/basic")
	if err != nil {
		t.Error(err)
	}

	err = upgrade.Write()
	if err != nil {
		t.Error(err)
	}
}
