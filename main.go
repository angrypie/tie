package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/angrypie/tie/tasks"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Creating microservices on top of golang packages (package as a service)"
	app.UsageText = "Use inside directory with tiel.yaml or let tie decide automaticaly"
	app.Action = defaultCommand

	//TODO log all temporary files to be able to remove them with clean command.
	app.Commands = []cli.Command{
		{
			Name:   "clean",
			Usage:  "Clean binaries",
			Action: cleanCommand,
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "gen",
			Usage: "set true to generate code without build and clean",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func defaultCommand(c *cli.Context) error {
	err := tasks.ReadConfigFile(".", c.Bool("gen"))
	if err != nil {
		if err != tasks.ErrConfigNotFound {
			return err
		}
		fmt.Println("Can't find tie.yaml in current directory")
		err := tasks.ReadDirAsConfig(".", c.Bool("gen"))
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
