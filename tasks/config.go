package tasks

import (
	"errors"
	"fmt"
	"go/build"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/angrypie/tie/types"
	"github.com/angrypie/tie/upgrade"
	"github.com/spf13/afero"

	yaml "gopkg.in/yaml.v2"
)

var ErrConfigNotFound = errors.New("config not found")

//ReadConfigFile trying to find tie.yml in specified direcotry
func ReadConfigFile(dest string) error {
	fs := afero.NewOsFs()
	configPath := fmt.Sprintf("%s/tie.yml", dest)
	buf, err := afero.ReadFile(fs, configPath)
	if err != nil {
		return ErrConfigNotFound
	}

	return configFromYaml(buf, dest)
}

func ReadDirAsConfig(dest string) error {
	fs := afero.NewOsFs()
	files, err := afero.ReadDir(fs, dest)
	if err != nil {
		return err
	}

	destPath, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	config := types.ConfigFile{
		Path: destPath,
	}

	basePath := strings.TrimPrefix(destPath, build.Default.GOPATH+"/src/")

	for _, file := range files {
		if file.IsDir() {
			pkgName := file.Name()
			//TODO why libs ignored?
			if strings.HasPrefix(pkgName, ".") || pkgName == "libs" {
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

			config.Services = append(config.Services, types.Service{
				Name: fmt.Sprintf("%s/%s", basePath, pkgName),
			})
			fmt.Println("Package added to config:", pkgName)
		}
	}

	if len(config.Services) == 0 {
		config.Services = append(config.Services, types.Service{Name: basePath, Type: "http"})
	}

	return withConfigFile(&config)
}

//Config execut different task based on tie.yml configurations
func configFromYaml(config []byte, dest string) (err error) {

	c := &types.ConfigFile{}
	err = yaml.Unmarshal(config, c)
	if err != nil {
		return err
	}

	//Default build path is tie.yml direcotry
	if c.Path == "" {
		destPath, err := filepath.Abs(dest)
		log.Println(dest, destPath)
		if err != nil {
			return err
		}
		c.Path = destPath
	}

	return withConfigFile(c)
}

func withConfigFile(c *types.ConfigFile) (err error) {
	var upgraders []*upgrade.Upgrader

	//Create upgraders
	for _, service := range c.Services {
		upgrader, err := upgradeWithServices(service, c.Services)
		if err != nil {
			return err
		}
		defer func() {
			err := upgrader.Clean()
			if err != nil {
				fmt.Println("Failed to clean upgrader", err)
			}
		}()
		upgraders = append(upgraders, upgrader)
	}

	cleanGoMod, err := initGoModules(c.Path)
	if err != nil {
		return
	}
	defer cleanGoMod()

	//Build upgraders
	for _, upgrader := range upgraders {
		err := upgrader.BuildTo(c.Path)
		if err != nil {
			return err
		}
	}

	return
}

//initGoModules initialize go module and return callback to clean changes after.
func initGoModules(dest string) (clean func(), err error) {
	clean = func() {}

	fs := afero.NewOsFs()
	goModulePath := fmt.Sprintf("%s/go.mod", dest)
	goModExist, err := afero.Exists(fs, goModulePath)
	if err != nil {
		return
	}

	if goModExist {
		clean = func() {
			output, err := exec.Command("sh", "-c", "go mod tidy").CombinedOutput()
			if err != nil {
				log.Println(string(output))
			}
		}
		return
	}

	clean = func() {
		goSumPath := fmt.Sprintf("%s/go.sum", dest)
		logError(fs.Remove(goModulePath), "removing go.mod")
		logError(fs.Remove(goSumPath), "removing go.sum")
	}

	output, err := exec.Command("sh", "-c", "go mod init").CombinedOutput()
	if err != nil {
		log.Println(string(output))
	}

	return
}

func logError(err error, msg string) {
	if err != nil {
		log.Println("ERR ", msg, err)
	}
}

//upgradeWithServices crate new upgrader for pkg and upgrade with services
func upgradeWithServices(current types.Service, services []types.Service) (*upgrade.Upgrader, error) {
	upgrader := upgrade.NewUpgrader(current)

	imports := make([]string, len(services))
	for i, service := range services {
		imports[i] = service.Name
	}

	err := upgrader.Upgrade(imports)
	if err != nil {
		return nil, err
	}

	return upgrader, nil
}
