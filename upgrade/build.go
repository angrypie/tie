package upgrade

import (
	"fmt"
	"log"
	"os/exec"
)

func (upgrader *Upgrader) Build() error {
	return upgrader.BuildTo("..")
}

//Build upgraded package binary to specified directory.
//Build source will be tie_upgraded if target is main package or tie_server for any other cases
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
