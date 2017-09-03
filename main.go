package main

import (
	"fmt"
	"os"

	"github.com/angrypie/tie/tasks"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Action = func(c *cli.Context) error {
		//TODO try to find config or try to use directory schema as config
		err := tasks.ReadConfigFile("./tie.yml")
		fmt.Println("Error:", err)
		return err
	}

	app.Run(os.Args)
}
