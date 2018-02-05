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

	app.Action = func(c *cli.Context) error {
		//TODO try to find config or try to use directory schema as config
		err := tasks.ReadConfigFile("./tie.yml")
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

	app.Run(os.Args)
}
