// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	u "os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vdemeester/praetorian/ssh"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup user name",
	Short: "Setup praetorian for the given user, with the given name.",
	Long: `Setup praetorian for the given user, with the given name.

One user (represented by its ssh public key) can have multiple name.`,

	Run: func(cmd *cobra.Command, args []string) {
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
			os.Exit(1)
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
		sshMagicString := fmt.Sprintf(`command="FINGERPRINT=%s NAME=%s praetorian exec '$SSH_ORIGINAL_COMMAND' && $SSH_ORIGINAL_COMMAND",no-X11-forwarding %s`, keyFingerPrint, name, finalKey)

		f, err := os.OpenFile(sshConfFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			// FIXME
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}

		defer f.Close()

		if _, err = f.WriteString(sshMagicString); err != nil {
			// FIXME
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
