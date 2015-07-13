package ssh

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// PublicSSHKey Struct that holds public ssh key informations
type PublicSSHKey struct {
	username string
	keyFile  *os.File
	content  string
}

// NewPublicSSHKey Create a new PublicSSHKey from a username and content
func NewPublicSSHKey(username, content string) *PublicSSHKey {
	return &PublicSSHKey{
		username: username,
		content:  content,
	}
}

// FingerPrint Get the fingerprint of a public ssh key
func (key *PublicSSHKey) FingerPrint() (string, error) {
	if key.keyFile == nil {
		if _, err := key.WriteToTemp(); err != nil {
			return "", err
		}
	}
	cmd := exec.Command("ssh-keygen", "-lf", key.keyFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error while getting ssh fingerprint : %s, %s (%s)", err, output, strings.Join(cmd.Args, " "))
	}
	parts := strings.SplitN(string(output), " ", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("Error while getting ssh fingerprint : %s, %s (%s)", err, output, strings.Join(cmd.Args, " "))
	}
	return parts[1], nil
}

func (key *PublicSSHKey) WriteToTemp() (*os.File, error) {
	if key.username == "" {
		return nil, fmt.Errorf("username cannot be nil or empty")
	}
	// Create a temporary file
	keyFile, err := ioutil.TempFile("", key.username)
	if err != nil {
		return nil, fmt.Errorf("Error %v", err)
	}
	ioutil.WriteFile(keyFile.Name(), []byte(key.content), 0600)
	key.keyFile = keyFile
	return keyFile, nil
}
