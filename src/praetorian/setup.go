package praetorian

import (
	"fmt"
	"os"
	e "os/exec"
	u "os/user"

	"github.com/codegangsta/cli"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var SetupCommand = cli.Command{
	Name:   "setup",
	Usage:  "Setup praetorian for the given user",
	Action: setup,
}

func setup(c *cli.Context) {
	if len(c.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "setup commands needs two arguments : praetorian setup user name\n")
		os.Exit(1)
	}
	username := c.Args().Get(0)
	name := c.Args().Get(1)

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
	// Create a temporary file
	keyFile, err := ioutil.TempFile("", username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %v", err)
		os.Exit(1)
	}
	ioutil.WriteFile(keyFile.Name(), key, 0600)

	// Use ssh-keygen
	cmd := e.Command("ssh-keygen", "-lf", keyFile.Name())
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting ssh fingerprint : %s, %s (%s)", err, output, strings.Join(cmd.Args, " "))
		os.Exit(1)
	}
	parts := strings.SplitN(string(output), " ", 3)
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "Error while getting ssh fingerprint : %s, %s (%s)", err, output, strings.Join(cmd.Args, " "))
		os.Exit(1)
	}
	keyFingerPrint := parts[1]

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
}
