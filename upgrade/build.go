package upgrade

import (
	"fmt"
	"os/exec"
)

func (upgrade *ServerUpgrade) Build() error {
	return upgrade.BuildTo("..")
}

func (upgrade *ServerUpgrade) BuildTo(dest string) error {
	path := upgrade.Package.Path + "/tie_server"
	alias := upgrade.Package.Alias
	buildComand := fmt.Sprintf(
		"cd %s && go build -o %s",
		path,
		dest+"/"+alias,
	)

	err := exec.Command("bash", "-c", buildComand).Run()
	if err != nil {
		return err
	}
	return nil
}

func (upgrade *ClientUpgrade) Build() error {
	return upgrade.BuildTo("..")
}

func (upgrade *ClientUpgrade) BuildTo(dist string) error {
	path := upgrade.Parser.Package.Path + "/tie_bin"
	alias := upgrade.Parser.Package.Alias
	buildComand := fmt.Sprintf(
		"cd %s && go build -o %s",
		path,
		dist+"/"+alias,
	)

	err := exec.Command("bash", "-c", buildComand).Run()
	if err != nil {
		return err
	}
	return nil
}
