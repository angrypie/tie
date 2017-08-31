package tasks

import (
	"log"

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
}

//Config execut different task based on tie.yml configurations
func Config(config []byte) (err error) {

	c := &ConfigFile{}
	err = yaml.Unmarshal(config, c)
	if err != nil {
		return err
	}

	var upgraders []*upgrade.Upgrader

	//Create upgraders and replace imports
	for _, service := range c.Services {
		//TODO rename replace function
		upgrader, err := Replace(service.Name, c.Services)
		if err != nil {
			return err
		}
		log.Println(service.Name)
		upgraders = append(upgraders, upgrader)
	}

	//Build upgraders
	for _, upgrader := range upgraders {
		err := upgrader.Build()
		if err != nil {
			return err
		}
	}

	//Clean tie_ folders
	for _, upgrader := range upgraders {
		upgrader.Clean()
		if err != nil {
			return err
		}
	}

	return nil
}
