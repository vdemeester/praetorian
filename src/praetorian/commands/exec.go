package commands

import (
	"bufio"
	"fmt"
	"os"
	e "os/exec"
	u "os/user"
	"path/filepath"
	"strings"
)

// ExecCommand a command, it is the wrapper
type ExecCommand struct {
	Meta
}

// Run The exec command
func (c *ExecCommand) Run(args []string) int {
	// Environment variable set in .authorized_keys
	// SSH_ORIGINAL_COMMAND
	sshOriginalCommand := os.Getenv("SSH_ORIGINAL_COMMAND")
	// Alias
	name := os.Getenv("NAME")
	// CONFIG FILE
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		user, _ := u.Current()
		configFile = filepath.Join(user.HomeDir, ".ssh", "praetorian")
	}
	fileInfo, err := os.Stat(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No praetorian configuration file, you won't be able to do anything")
		os.Exit(1)
	}
	if fileInfo.Mode() != 0600 {
		fmt.Fprintf(os.Stderr, "Praetorian file should be only readable by you (mode 0600)\nAborting..")
		os.Exit(1)
	}

	// Source the config file (old behaviour for now)
	allowedCommands, err := parseConfigurationFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while loading configuration file %s\nAborting..", configFile)
		os.Exit(1)
	}
	parts := strings.SplitN(sshOriginalCommand, " ", 2)
	command := parts[0]
	sshargs := []string{}
	if len(parts) == 2 {
		sshargs = strings.Split(parts[1], " ")
	}

	allowed := false

	fmt.Printf("%s in %v", command, allowedCommands)
	for _, allowedCommand := range allowedCommands[name] {
		if command == allowedCommand {
			allowed = true
			break
		}
	}

	if allowed {
		cmd := e.Command(command, sshargs...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error while running : %s %s", command, strings.Join(sshargs, " "))
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Alias %s Invalid command %s", name, sshOriginalCommand)
		os.Exit(1)
	}
	return 0
}

// Synopsis is a one-line, short synopsis of the command.
func (c *ExecCommand) Synopsis() string {
	return "Try to execute a command"
}

// Help is a long-form help text that includes the command-line
// usage, a brief few sentences explaining the function of the command,
// and the complete list of flags the command accepts.
func (c *ExecCommand) Help() string {
	helpText := `
Usage: praetorian exec commands [args]
  Try to execute a command. 
`
	return strings.TrimSpace(helpText)
}

func parseConfigurationFile(filename string) (map[string][]string, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return map[string][]string{}, err
	}
	defer fh.Close()

	lines := map[string][]string{}
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		// line is not empty, and not starting with '#'
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			if strings.Contains(line, "=") {
				data := strings.SplitN(line, "=", 2)

				// trim the front of a variable, but nothing else
				variable := strings.TrimLeft(data[0], whiteSpaces)
				if !strings.ContainsAny(variable, whiteSpaces) {
					// pass the value through, no trimming
					value := strings.Replace(data[1], `"`, "", -1)
					lines[variable] = strings.Split(value, " ")
				}
			}
		}
	}
	return lines, scanner.Err()
}

var whiteSpaces = " \t"
