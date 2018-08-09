package main

import (
	"fmt"
	"log"
	"os"

	"github.com/angrypie/tie/tasks"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Creating microservices on top of golang packages (package as a service)"
	app.UsageText = "Use inside directory with tiel.yml or let tie decide automaticaly"
	app.Action = defaultCommand

	app.Commands = []cli.Command{
		{
			Name:   "clean",
			Usage:  "Clean binaries",
			Action: cleanCommand,
		},
	}

	app.Run(os.Args)
}

func defaultCommand(c *cli.Context) error {
	err := tasks.ReadConfigFile(".")
	if err != nil {
		if err.Error() != "Cant read file" {
			log.Println(err)
			return err
		}
		fmt.Println("Can't find tie.yml in current directory")
		err := tasks.ReadDirAsConfig(".")
		if err != nil {
			fmt.Println(err)
		}
	}

	return err
}

func cleanCommand(c *cli.Context) error {
	err := tasks.CleanBinary(".")
	return err
}
