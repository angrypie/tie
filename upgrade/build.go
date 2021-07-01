package upgrade

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/afero"
)

//BuildTo builds upgraded package binary to specified directory.
//For main packages build source dir is tie_modules/upgraded.
func (upgrader *Upgrader) BuildTo(dist string) error {
	alias := upgrader.Parser.Service.Alias
	buildDir := "tie_modules"
	if upgrader.Parser.GetPackageName() == "main" {
		buildDir = "tie_modules/upgraded"
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
			return errors.New("directory with same name as binary exist")
		}
	}

	buildComand := fmt.Sprintf(
		"cd %s && go mod tidy && go build -o %s/%s",
		path,
		dist,
		binName,
	)
	fmt.Println("Build command:", buildComand)

	output, err := exec.Command("sh", "-c", buildComand).CombinedOutput()
	if err != nil {
		fmt.Println("Build erorr: ", string(output))
		return err
	}

	return nil
}
