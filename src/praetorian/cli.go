package praetorian

import (
	"fmt"
	"os"
	"praetorian/commands"

	"github.com/codegangsta/cli"
)

// Application error
type AppError struct {
	message  string
	exitCode int
}

func (e AppError) Error() string {
	return e.message
}

// Run execute RunCustom() with color and output to Stdout/Stderr.
// It returns exit code.
func Run(args []string) int {
	var exitCode = 0

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Usage = Description
	app.Commands = []cli.Command{
		commands.ExecCommand,
		commands.SetupCommand,
	}

	err := app.Run(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute: %s\n", err.Error())
	}
	if err, ok := err.(AppError); ok {
		exitCode = err.exitCode
	}

	return exitCode
}
