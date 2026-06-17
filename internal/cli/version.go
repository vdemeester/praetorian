package cli

import (
	"fmt"

	"github.com/vdemeester/praetorian/version"
)

func versionCmd(_ []string) int {
	fmt.Printf("praetorian %s (commit %s, built %s)\n", version.Version, version.Commit, version.Date)
	return 0
}
