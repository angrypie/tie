package tasks

import (
	"errors"
	"fmt"
	"go/build"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

//ReadConfigFile trying to find tie.yml in specified direcotry
func ReadConfigFile(dest string) error {
	fs := afero.NewOsFs()
	configPath := fmt.Sprintf("%s/tie.yml", dest)
	buf, err := afero.ReadFile(fs, configPath)
	if err != nil {
		return errors.New("Cant read file")
	}

	return ConfigFromYaml(buf, dest)
}

func ReadDirAsConfig(dest string) error {
	fs := afero.NewOsFs()
	files, err := afero.ReadDir(fs, dest)
	if err != nil {
		return err
	}

	config := ConfigFile{}

	destPath, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	basePath := strings.TrimPrefix(destPath, build.Default.GOPATH+"/src/")

	for _, file := range files {
		if file.IsDir() {
			pkgName := file.Name()
			if strings.HasPrefix(pkgName, ".") {
				continue
			}

			rfs := afero.NewRegexpFs(fs, regexp.MustCompile(`\.go$`))

			goFiles, err := afero.ReadDir(rfs, fmt.Sprintf("%s/%s", dest, pkgName))
			if err != nil {
				return err
			}

			//TODO file with .go extension should not be directories
			if len(goFiles) == 0 {
				fmt.Println("Folder ignored:", pkgName)
				continue
			}

			config.Services = append(config.Services, Service{
				Name: fmt.Sprintf("%s/%s", basePath, pkgName),
			})
			fmt.Println("Package added to config:", pkgName)
		}
	}

	return Config(&config)
}
