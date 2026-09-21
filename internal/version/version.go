package version

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Info() string {
	if Commit == "unknown" && Date == "unknown" {
		return Version
	}
	return Version + " (" + Commit + ", " + Date + ")"
}
