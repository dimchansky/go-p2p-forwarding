package cmd

import "fmt"

var (
	// Version (set by compiler) is the version of program
	Version = "undefined"
	// BuildTime (set by compiler) is the program build time in '+%Y-%m-%dT%H:%M:%SZ' format
	BuildTime = "undefined"
	// GitHash (set by compiler) is the git commit hash of source tree
	GitHash = "undefined"
)

// PrintVersion outputs the binary version
func PrintVersion() {
	fmt.Printf("Version: %s\nBuildTime: %v\nGitHash: %s\n", Version, BuildTime, GitHash)
}
