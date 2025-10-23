// Package version provides version information for the outfit picker application.
package version

import "fmt"

// Build-time variables
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// GetVersion returns the full version string
func GetVersion() string {
	if Commit != "unknown" && Date != "unknown" {
		return fmt.Sprintf("%s (%s, built %s)", Version, Commit[:7], Date)
	}
	return Version
}
