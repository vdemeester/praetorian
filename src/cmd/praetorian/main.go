package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "praetorian"
	app.Usage = "A ssh praetorian (bouncer, minder or whatever) ; it's just a cool restricted command script."
	app.Action = func(c *cli.Context) {
		println("Hello friend!")
	}

	app.Run(os.Args)
}
