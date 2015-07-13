package commands

import (
	"fmt"
	"os"
	u "os/user"

	"io/ioutil"
	"path/filepath"
	"strings"

	"praetorian/ssh"
)

// SetupCommand command that will setup the ssh magic
type SetupCommand struct {
	Meta
}

// Run The setup command
func (c *SetupCommand) Run(args []string) int {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "setup commands needs two arguments : praetorian setup user name\n")
		os.Exit(1)
	}
	username := args[0]
	name := args[1]

	// Does the user exists
	user, err := u.Lookup(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "User %s does not exists, aborting.\n", username)
		os.Exit(1)
	}

	// Read key from stdin
	key, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin\n")
		os.Exit(1)
	}

	// Let's get the fingerprint of the key
	sshKey := ssh.NewPublicSSHKey(username, string(key))
	keyFingerPrint, err := sshKey.FingerPrint()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	sshConfDir := filepath.Join(user.HomeDir, ".ssh")
	sshConfFile := filepath.Join(sshConfDir, "authorized_keys")
	// Does $HOME/.ssh exists
	if _, err := os.Stat(sshConfDir); err != nil {
		os.MkdirAll(sshConfDir, 0600)
	}
	// Does $HOME/.ssh/authorized_keys exists
	if _, err := os.Stat(sshConfFile); err != nil {
		ioutil.WriteFile(sshConfFile, []byte{}, 0600)
	}

	// Put the magic string at the end of sshConfFile
	finalKey := strings.TrimSpace(string(key))
	sshMagicString := fmt.Sprintf(`%s command="FINGERPRINT=%s NAME=%s praetorian exec $SSH_ORIGINAL_COMMAND",no-X11-forwarding`, finalKey, keyFingerPrint, name)

	f, err := os.OpenFile(sshConfFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		// FIXME
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(sshMagicString); err != nil {
		// FIXME
		panic(err)
	}
	return 0
}

// Synopsis is a one-line, short synopsis of the command.
func (c *SetupCommand) Synopsis() string {
	return "Setup praetorian for the given user"
}

// Help is a long-form help text that includes the command-line
// usage, a brief few sentences explaining the function of the command,
// and the complete list of flags the command accepts.
func (c *SetupCommand) Help() string {
	helpText := `
Usage: praetorian setup user name
  Setup praetorian for the given user, with the given name. 
  One user can have multiple name.
`
	return strings.TrimSpace(helpText)
}
