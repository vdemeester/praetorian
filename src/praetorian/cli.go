package praetorian

import (
	"os"
	"praetorian/commands"

	"github.com/mitchellh/cli"
	"log"
)

// Run execute RunCustom() with color and output to Stdout/Stderr.
// It returns exit code.
func Run(args []string) int {
	meta := &commands.Meta{
		UI: &cli.ColoredUi{
			InfoColor:  cli.UiColorBlue,
			ErrorColor: cli.UiColorRed,
			Ui: &cli.BasicUi{
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
				Reader:      os.Stdin,
			},
		}}

	c := cli.NewCLI(Name, Version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"exec": func() (cli.Command, error) {
			return &commands.ExecCommand{
				Meta: *meta,
			}, nil
		},
		"setup": func() (cli.Command, error) {
			return &commands.SetupCommand{
				Meta: *meta,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	return exitStatus
}
