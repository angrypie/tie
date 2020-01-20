package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/angrypie/tie/tasks"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Creating microservices on top of golang packages (package as a service)"
	app.UsageText = "Use inside directory with tiel.yml or let tie decide automaticaly"
	app.Action = defaultCommand

	app.Commands = []*cli.Command{
		{
			Name:   "clean",
			Usage:  "Clean binaries",
			Action: cleanCommand,
		},
		{
			Name:   "install-deps",
			Usage:  "Insall Tie dependencies",
			Action: installGoDependencies,
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
	removed, err := tasks.CleanBinary(".")
	if err != nil {
		return err
	}
	if length := len(removed); length != 0 {
		fmt.Printf("Deleted %d files: %s\n", len(removed), strings.Join(removed, ", "))
	} else {
		fmt.Printf("Nothing to clean\n")
	}
	return nil
}

func installGoDependencies(c *cli.Context) error {
	err := tasks.InstallGoDependencies()
	if err != nil {
		return err
	}
	fmt.Printf("Dependencies installed\n")
	return nil
}
