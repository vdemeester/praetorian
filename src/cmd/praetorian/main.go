package main

import (
	"os"
	"praetorian"

	"github.com/codegangsta/cli"
)

const (
	// NAME name of the application
	NAME = "praetorian"
	// VERSION version of the application
	VERSION = "0.5.0-dev"
	// DESCRIPTION description of the application
	DESCRIPTION = "A ssh praetorian (bouncer, minder or whatever) ; it's just a cool restricted command script."
)

func main() {
	app := cli.NewApp()
	app.Name = NAME
	app.Version = VERSION
	app.Usage = DESCRIPTION
	app.Commands = []cli.Command{
		{
			Name:   "exec",
			Usage:  "Try to execute a command",
			Action: praetorian.Exec,
		},
		{
			Name:   "setup",
			Usage:  "Setup praetorian for the given user",
			Action: praetorian.Setup,
		},
	}

	app.Run(os.Args)
}
