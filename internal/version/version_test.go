package version

import "testing"

func TestInfoDev(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	t.Cleanup(func() {
		Version, Commit, Date = oldVersion, oldCommit, oldDate
	})

	Version = "dev"
	Commit = "unknown"
	Date = "unknown"

	if got, want := Info(), "dev"; got != want {
		t.Fatalf("Info() = %q, want %q", got, want)
	}
}

func TestInfoBuildMetadata(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	t.Cleanup(func() {
		Version, Commit, Date = oldVersion, oldCommit, oldDate
	})

	Version = "1.2.3"
	Commit = "abc123"
	Date = "2026-09-21"

	if got, want := Info(), "1.2.3 (abc123, 2026-09-21)"; got != want {
		t.Fatalf("Info() = %q, want %q", got, want)
	}
}
