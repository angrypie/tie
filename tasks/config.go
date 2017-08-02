package tasks

import (
	"log"

	yaml "gopkg.in/yaml.v2"
)

type Service struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
	Type  string `yaml:"type"`
}

type ConfigFile struct {
	Services []Service `yaml:"services"`
}

//Config execut different task based on tie.yml configurations
func Config(config []byte) (err error) {

	c := &ConfigFile{}
	err = yaml.Unmarshal(config, c)
	if err != nil {
		return err
	}

	for _, service := range c.Services {
		err := Binary(service.Name, "/tmp/tie", c.Services)
		if err != nil {
			return err
		}
		log.Println(service.Name)
	}

	return nil
}
