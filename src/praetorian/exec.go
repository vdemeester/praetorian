package praetorian

import (
	"github.com/codegangsta/cli"
)

var ExecCommand = cli.Command{
	Name:   "exec",
	Usage:  "Try to execute a command",
	Action: exec,
}

// Exec a command, it is the wrapper
func exec(c *cli.Context) {
	// Environment variable set in .authorized_keys
	// SSH_ORIGINAL_COMMAND
	// USER
	// CONFIG FILE
}
