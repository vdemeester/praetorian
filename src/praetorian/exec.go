package praetorian

import (
	"github.com/codegangsta/cli"
)

// Exec a command, it is the wrapper
func Exec(c *cli.Context) {
	// Environment variable set in .authorized_keys
	// SSH_ORIGINAL_COMMAND
	// USER
	// CONFIG FILE
}
