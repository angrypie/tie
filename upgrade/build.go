package upgrade

import (
	"fmt"
	"log"
	"os/exec"
)

func (upgrader *Upgrader) Build() error {
	return upgrader.BuildTo("..")
}

func (upgrader *Upgrader) BuildTo(dist string) error {
	alias := upgrader.Parser.Package.Alias
	log.Println(alias)
	buildDir := "tie_server"
	if upgrader.Parser.GetPackageName() == "main" {
		buildDir = "tie_upgraded"
	}

	path := fmt.Sprintf("%s/%s", upgrader.Parser.Package.Path, buildDir)
	buildComand := fmt.Sprintf(
		"cd %s && go build -o %s",
		path,
		dist+"/"+alias,
	)

	log.Println(buildComand)
	err := exec.Command("bash", "-c", buildComand).Run()
	if err != nil {
		return err
	}
	return nil
}
