package tasks

import (
	"fmt"
	"go/build"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

func ReadConfigFile(dest string) error {
	fs := afero.NewOsFs()
	buf, err := afero.ReadFile(fs, dest)
	if err != nil {
		return err
	}

	return ConfigFromYaml(buf)
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

			rfs := afero.NewRegexpFs(fs, regexp.MustCompile(`\.go$`))

			goFiles, err := afero.ReadDir(rfs, fmt.Sprintf("%s/%s", dest, pkgName))
			if err != nil {
				return err
			}

			//TODO file with .go extension should not be directoies
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
