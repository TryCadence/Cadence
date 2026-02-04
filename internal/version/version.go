package version

var (
	Version   = "unknown"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func String() string {
	return Version
}

func Full() string {
	str := "Cadence " + Version
	if GitCommit != "unknown" && GitCommit != "" {
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
