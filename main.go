// Command praetorian is an SSH command restrictor. It is used as the target of
// a `command="praetorian run <alias>"` directive in authorized_keys, and
// validates SSH_ORIGINAL_COMMAND against an allow-list before executing it
// directly (no shell).
package main

import (
	"os"

	"github.com/vdemeester/praetorian/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
