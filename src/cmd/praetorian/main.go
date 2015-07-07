package main

import (
	"os"
	"praetorian"
	commands "praetorian/commands"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = praetorian.Name
	app.Version = praetorian.Version
	app.Usage = praetorian.Description
	app.Commands = []cli.Command{
		commands.ExecCommand,
		commands.SetupCommand,
	}

	app.Run(os.Args)
}
