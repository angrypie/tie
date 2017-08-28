package upgrade

import "testing"

func TestUpgrade(t *testing.T) {
	server, err := Server("github.com/angrypie/tie/example/basic")
	if err != nil {
		t.Error(err)
	}

	client, err := Client("github.com/angrypie/tie/example/basic/usage")
	if err != nil {
		t.Error(err)
	}
	ok := client.Replace(
		"github.com/angrypie/tie/example/basic",
		"github.com/angrypie/tie/example/basic/tie_client",
	)
	if !ok {
		t.Error("Imports not replaced successfuly")
	}

	err = server.Write()
	if err != nil {
		t.Error(err)
	}

	err = client.Write()
	if err != nil {
		t.Error(err)
	}

	err = server.BuildTo("/tmp")
	if err != nil {
		t.Error(err)
	}

	err = client.BuildTo("/tmp")
	if err != nil {
		t.Error(err)
	}

	err = server.Clean()
	if err != nil {
		t.Error(err)
	}

	err = client.Clean()
	if err != nil {
		t.Error(err)
	}
}

func TestUpgrader(t *testing.T) {
	upgrader := NewUpgrader("github.com/angrypie/tie/example/basic")
	err := upgrader.Parse()
	if err != nil {
		t.Error(err)
	}

	err = upgrader.Make()
	if err != nil {
		t.Error(err)
	}

	err = upgrader.Write()
	if err != nil {
		t.Error(err)
	}

	upgrader = NewUpgrader("github.com/angrypie/tie/example/basic/usage")

	err = upgrader.Parse()
	if err != nil {
		t.Error(err)
	}

	ok := upgrader.Replace([]string{"github.com/angrypie/tie/example/basic"})
	if !ok {
		t.Error("Imports should be replaced")
	}

	err = upgrader.Make()
	if err != nil {
		t.Error(err)
	}

	err = upgrader.Write()
	if err != nil {
		t.Error(err)
	}
}
