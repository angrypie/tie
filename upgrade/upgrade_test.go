package upgrade

import "testing"

func TestUpgrader(t *testing.T) {
	e := func(err error) {
		if err != nil {
			t.Error(err)
		}
	}

	//Upgrade basic example
	basic := NewUpgrader("github.com/angrypie/tie/example/basic")
	e(basic.Parse())

	e(basic.Make())

	e(basic.Write())

	//Upgrade basic/usage example
	usage := NewUpgrader("github.com/angrypie/tie/example/basic/usage")
	e(usage.Parse())

	e(usage.Replace([]string{"github.com/angrypie/tie/example/basic"}))

	e(usage.Make())
	e(usage.Write())

	//Build basic and usage
	e(usage.Build())
	e(basic.Build())

	//Clean basic and usage
	//e(usage.Clean())
	//e(basic.Clean())

}
