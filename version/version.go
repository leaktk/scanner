package version

import "fmt"

// Version number set by the build
var Version = ""

// Commit id set by the build
var Commit = ""

func PrintVersion() {
	if len(Version) > 0 {
		fmt.Printf("Version: %v\n", Version)

		if len(Commit) > 0 {
			fmt.Printf("Commit: %v\n", Commit)
		}
	} else {
		fmt.Println("Version information not available")
	}
}
