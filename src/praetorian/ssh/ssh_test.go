package ssh

import (
	"io/ioutil"
	"testing"
)

func TestPublicSSHKeyWriteToTempErrors(t *testing.T) {
	invalid := []PublicSSHKey{
		PublicSSHKey{
			content: "some content",
		},
		PublicSSHKey{
			username: "",
			content:  "some content",
		},
	}
	for _, sshKey := range invalid {
		if _, err := sshKey.WriteToTemp(); err == nil {
			t.Fatalf("Should have returned an error, got nothing")
		}
	}
}

func TestPublicSSHKeyWriteToTemp(t *testing.T) {
	sshKey := &PublicSSHKey{
		username: "somebody",
		content:  "some content",
	}
	file, err := sshKey.WriteToTemp()
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "some content" {
		t.Fatalf("Expected the written file to have 'some content' as content, got %v", string(data))
	}
}

func TestPublicSSHKeyFingerPrintErrors(t *testing.T) {
	invalid := []PublicSSHKey{
		PublicSSHKey{
			content: "some content",
		},
		PublicSSHKey{
			username: "username",
			content:  "some content",
		},
	}
	for _, sshKey := range invalid {
		if _, err := sshKey.FingerPrint(); err == nil {
			t.Fatalf("Should have returned an error, got nothing")
		}
	}
}

func TestPublicSSHKeyFingerPrint(t *testing.T) {
	expectedFingerPrint := "SHA256:pyIviSnX1wCz//lp7kkixlk/1GJNUafzrCwBGMqe3ZI"
	sshKey := &PublicSSHKey{
		username: "somebody",
		content:  `ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSUGPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XAt3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/EnmZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbxNrRFi9wrf+M7Q== schacon@mylaptop.local`,
	}
	fingerPrint, err := sshKey.FingerPrint()
	if err != nil {
		t.Fatal(err)
	}
	if fingerPrint != expectedFingerPrint {
		t.Fatalf("Expected %s, got %s", expectedFingerPrint, fingerPrint)
	}

}
