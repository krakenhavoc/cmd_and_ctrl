package main

import (
	"bytes"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

// --version prints the stamped version and commit and starts no seat: a
// seat would block on stdin for MCP frames, so returning at all proves it.
func TestVersionFlagPrintsVersionAndCommit(t *testing.T) {
	oldV, oldC := version, commit
	t.Cleanup(func() { version, commit = oldV, oldC })
	version, commit = "v1.2.3", "0123456789abcdef"

	for _, arg := range []string{"--version", "-version"} {
		var out bytes.Buffer
		if err := run([]string{arg}, &out); err != nil {
			t.Fatalf("run(%s): %v", arg, err)
		}
		got := strings.TrimSpace(out.String())
		want := "mcpseat v1.2.3 (commit 0123456789abcdef, " + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH + ")"
		if got != want {
			t.Errorf("run(%s) printed %q, want %q", arg, got, want)
		}
	}
}

func TestBuildIdentity(t *testing.T) {
	withVCS := &debug.BuildInfo{
		Main:     debug.Module{Version: "v0.4.0"},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abc123"}},
	}
	devel := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}

	cases := []struct {
		name         string
		ver, rev     string
		bi           *debug.BuildInfo
		wantV, wantC string
	}{
		{"no build info keeps the defaults", "dev", "unknown", nil, "dev", "unknown"},
		{"a local build with no VCS keeps the defaults", "dev", "unknown", devel, "dev", "unknown"},
		{"go install and a git checkout fill both", "dev", "unknown", withVCS, "v0.4.0", "abc123"},
		{"stamped values win", "v9.9.9", "fff", withVCS, "v9.9.9", "fff"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v, r := buildIdentity(c.ver, c.rev, c.bi)
			if v != c.wantV || r != c.wantC {
				t.Errorf("buildIdentity = (%q, %q), want (%q, %q)", v, r, c.wantV, c.wantC)
			}
		})
	}
}
