package upgrade

import "testing"

func TestUpgrader(t *testing.T) {
	e := func(err error) {
		if err != nil {
			t.Error(err)
		}
	}

	//Upgrade basit example
	basic := NewUpgrader("github.com/angrypie/tie/example/custom-types/register")
	e(basic.Parse())
	e(basic.Make())
	e(basic.Write())

	//Upgrade basic/usage example
	usage := NewUpgrader("github.com/angrypie/tie/example/custom-types/cli")
	e(usage.Parse())
	e(usage.Replace([]string{"github.com/angrypie/tie/example/custom-types/register"}))
	e(usage.Write())

	//Build basic and usage
	//e(usage.Build())
	//e(basic.Build())

	//Clean basic and usage
	e(usage.Clean())
	e(basic.Clean())
}
