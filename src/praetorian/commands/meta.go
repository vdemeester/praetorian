package commands

import "github.com/mitchellh/cli"

// Meta contain the meta-option that nealy all subcommand inherited.
type Meta struct {
	UI cli.Ui
}
