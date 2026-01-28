package version

const (
	// Version is the current version of Cadence
	Version = "0.1.0"
	// BuildTime is the build timestamp (set at compile time with -ldflags)
	BuildTime = "unknown"
	// GitCommit is the git commit hash (set at compile time with -ldflags)
	GitCommit = "unknown"
)

// String returns the full version string
func String() string {
	return Version
}

// Full returns a detailed version string with build info
func Full() string {
	str := "Cadence v" + Version
	if GitCommit != "unknown" && len(GitCommit) > 0 {
		commit := GitCommit
		if len(commit) > 8 {
			commit = commit[:8]
		}
		str += " (" + commit + ")"
	}
	if BuildTime != "unknown" {
		str += " built at " + BuildTime
	}
	return str
}
