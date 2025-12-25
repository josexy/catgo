package version

import "fmt"

var (
	Version   = "unknown"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func Show() {
	fmt.Println("Version: " + Version)
	fmt.Println("Git Commit: " + GitCommit)
	fmt.Println("Build Time: " + BuildTime)
	fmt.Println("Go Version: " + GoVersion)
}
