package tasks

import (
	"log"
	"path/filepath"

	"github.com/angrypie/tie/upgrade"

	yaml "gopkg.in/yaml.v2"
)

type Service struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
	Type  string `yaml:"type"`
}

type ConfigFile struct {
	Services []Service `yaml:"services"`
	Path     string    `yaml:"path"`
}

//Config execut different task based on tie.yml configurations
func ConfigFromYaml(config []byte, dest string) (err error) {

	c := &ConfigFile{}
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

	return Config(c)
}

func Config(c *ConfigFile) error {
	var upgraders []*upgrade.Upgrader

	//Create upgraders and replace imports
	for _, service := range c.Services {
		upgrader, err := upgradeWithServices(service, c.Services)
		if err != nil {
			return err
		}
		defer upgrader.Clean()
		upgraders = append(upgraders, upgrader)
	}

	//Build upgraders
	for _, upgrader := range upgraders {
		err := upgrader.BuildTo(c.Path)
		if err != nil {
			return err
		}
	}

	//Clean tie_ folders
	for _, upgrader := range upgraders {
		err := upgrader.Clean()
		if err != nil {
			return err
		}
	}

	return nil
}
