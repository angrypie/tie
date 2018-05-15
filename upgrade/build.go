package upgrade

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/afero"
)

//Build calls BuildTo method with parent direcotry as argument
func (upgrader *Upgrader) Build() error {
	return upgrader.BuildTo(".")
}

//Build upgraded package binary to specified directory.
//Build source will be tie_upgraded if target is main package or tie_server for any other cases
func (upgrader *Upgrader) BuildTo(dist string) error {
	alias := upgrader.Parser.Service.Alias
	buildDir := "tie_server"
	if upgrader.Parser.GetPackageName() == "main" {
		buildDir = "tie_upgraded"
	}

	path := fmt.Sprintf("%s/%s", upgrader.Parser.Package.Path, buildDir)

	fs := afero.NewOsFs()
	binName := fmt.Sprintf("%s.run", alias)
	ok, err := afero.Exists(fs, fmt.Sprintf("%s/%s", path, dist+"/"+binName))
	if err != nil {
		return err
	}
	if ok {
		if ok, _ := afero.IsDir(fs, fmt.Sprintf("%s/%s", path, dist+"/"+binName)); ok {
			return errors.New("Directory with same name as binary exist")
		}
	}

	buildComand := fmt.Sprintf(
		"cd %s && go build -o %s/%s",
		path,
		dist,
		binName,
	)
	fmt.Println(buildComand)

	err = exec.Command("sh", "-c", buildComand).Run()
	if err != nil {
		return err
	}
	return nil
}
