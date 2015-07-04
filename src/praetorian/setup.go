package praetorian

import (
	"github.com/codegangsta/cli"
)

var SetupCommand = cli.Command{
	Name:   "setup",
	Usage:  "Setup praetorian for the given user",
	Action: setup,
}

func setup(c *cli.Context) {

}
