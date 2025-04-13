// Package version provides version information for the application
package version

import (
	"fmt"

	evm "github.com/BOTCoinNetwork/BVM/src/version"
	_babble "github.com/BOTCoinNetwork/babble/src/version"
	geth "github.com/ethereum/go-ethereum/params"
)

var (
	//Version is the full version string
	Version = "0.3.6"

	// GitCommit is set with --ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
	GitCommit string
	// GitBranch is set with --ldflags "-X main.gitBranch=$(git symbolic-ref --short HEAD)"
	GitBranch string
)

func init() {
	// branch is only of interest if it is not the master branch
	if GitBranch != "" && GitBranch != "master" {
		Version += "-" + GitBranch
	}

	if GitCommit != "" {
		Version += "-" + GitCommit[:8]
	}
}

// FullVersion outputs version information for Monet, EVM-Lite, Babble and Geth
func FullVersion() string {
	return fmt.Sprintln("Monetd Version: "+Version) +
		fmt.Sprintln("     EVM-Lite Version: "+evm.Version) +
		fmt.Sprintln("     Babble Version: "+_babble.Version) +
		fmt.Sprintln("     Geth Version: "+geth.Version)
}
