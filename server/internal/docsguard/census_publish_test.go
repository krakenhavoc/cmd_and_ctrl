package docsguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/coverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/roadmap"
)

// The census-publish job in ci-cd.yml pushes to develop and main with a
// deploy key that bypasses the branch rulesets, so what it may commit is
// the one security-relevant thing about it. Luke's condition on
// Discussion #1231: its check stays an allowlist of exact generated
// blocks. This test runs the check itself, the bash between the
// ">>> publish-allowlist" and "<<< publish-allowlist" comments, against
// a scratch repository holding the real docs, and proves it accepts a
// change confined to the two generated blocks (the catalog census and,
// since #1461, the Closed seams list) and refuses everything else.
//
// It also ties the workflow's marker prefixes to the Go constants that
// write the blocks: the "accept" cases are produced by the real splice
// functions, so a marker renamed on one side only fails here instead of
// failing every publish on develop.

const (
	workflowPath = "../../../.github/workflows/ci-cd.yml"
	censusDoc    = "docs/decklists/card-coverage-roadmap.md"
	seamsDoc     = "docs/engine-seams.md"
)

func publishAllowlistScript(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read %s: %v", workflowPath, err)
	}
	lines := strings.Split(string(raw), "\n")
	start, end := -1, -1
	for i, l := range lines {
		switch strings.TrimSpace(l) {
		case "# >>> publish-allowlist":
			if start >= 0 {
				t.Fatalf("%s has two >>> publish-allowlist markers", workflowPath)
			}
			start = i
		case "# <<< publish-allowlist":
			if end >= 0 {
				t.Fatalf("%s has two <<< publish-allowlist markers", workflowPath)
			}
			end = i
		}
	}
	if start < 0 || end < start {
		t.Fatalf("%s: census-publish's allowlist check is no longer fenced by "+
			"\"# >>> publish-allowlist\" / \"# <<< publish-allowlist\"; this test runs the code between them", workflowPath)
	}
	indent := lines[start][:len(lines[start])-len(strings.TrimLeft(lines[start], " "))]
	var b strings.Builder
	for _, l := range lines[start : end+1] {
		b.WriteString(strings.TrimPrefix(l, indent) + "\n")
	}
	return b.String()
}

// Every `git add` in the workflow stages the allowlisted files and
// nothing else, so a file outside the allowlist can never be committed
// even if the check above were bypassed.
func TestCensusPublishStagesOnlyTheAllowlist(t *testing.T) {
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read %s: %v", workflowPath, err)
	}
	adds := 0
	for _, l := range strings.Split(string(raw), "\n") {
		l = strings.TrimSpace(l)
		if !strings.HasPrefix(l, "git add") {
			continue
		}
		adds++
		if l != `git add -- "${ALLOW_FILES[@]}"` {
			t.Errorf("%s stages something other than the publish allowlist: %q", workflowPath, l)
		}
	}
	if adds == 0 {
		t.Errorf("%s: census-publish no longer stages the allowlist; update this test with it", workflowPath)
	}
}

func TestCensusPublishAllowlistAcceptsOnlyTheGeneratedBlocks(t *testing.T) {
	for _, tool := range []string{"bash", "git", "awk", "diff"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH: %v", tool, err)
		}
	}
	script := publishAllowlistScript(t)

	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join("../../..", rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		return string(raw)
	}
	census, seams := read(censusDoc), read(seamsDoc)

	newCensus := func(doc string) string {
		next, ok := coverage.Splice(doc, coverage.BeginMarker+"\n| a regenerated census | 1 |\n"+coverage.EndMarker)
		if !ok {
			t.Fatalf("%s has no census markers the coverage package recognises", censusDoc)
		}
		return next
	}
	newClosed := func(doc string) string {
		next, ok := roadmap.SpliceClosedSeams(doc, roadmap.ClosedSeamsBlock([]roadmap.ClosedSeam{
			{Title: "A new closure", Body: "**A new closure** (#1) — regenerated."},
		}))
		if !ok {
			t.Fatalf("%s has no closed-seams markers the roadmap package recognises", seamsDoc)
		}
		return next
	}
	newOpen := func(doc string) string {
		next, ok := roadmap.SpliceSeams(doc, roadmap.SeamsBeginMarker+"\n| hand edit |\n"+roadmap.SeamsEndMarker)
		if !ok {
			t.Fatalf("%s has no open-seams markers", seamsDoc)
		}
		return next
	}
	prose := func(doc string) string { return strings.Replace(doc, "\n", "\nAn extra line of prose.\n", 1) }

	cases := []struct {
		name   string
		edits  map[string]string
		accept bool
	}{
		{"nothing changed", nil, true},
		{"census block only", map[string]string{censusDoc: newCensus(census)}, true},
		{"closed-seams block only", map[string]string{seamsDoc: newClosed(seams)}, true},
		{"both generated blocks", map[string]string{censusDoc: newCensus(census), seamsDoc: newClosed(seams)}, true},

		{"the Open seams table, which a PR owns", map[string]string{seamsDoc: newOpen(seams)}, false},
		{"engine-seams prose", map[string]string{seamsDoc: prose(seams)}, false},
		{"closed block plus engine-seams prose", map[string]string{seamsDoc: prose(newClosed(seams))}, false},
		{"census prose", map[string]string{censusDoc: prose(census)}, false},
		{"census block plus census prose", map[string]string{censusDoc: prose(newCensus(census))}, false},
		{"another file", map[string]string{"AGENTS.md": "changed\n"}, false},
		{"both blocks plus another file", map[string]string{
			censusDoc: newCensus(census), seamsDoc: newClosed(seams), "AGENTS.md": "changed\n",
		}, false},
		{"a fragment", map[string]string{"docs/engine-seams/closed/1-x.md": "changed\n"}, false},
		{"closed block left unterminated", map[string]string{
			seamsDoc: strings.Replace(seams, roadmap.ClosedEndMarker, "anything at all", 1),
		}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{
				censusDoc:                         census,
				seamsDoc:                          seams,
				"AGENTS.md":                       "original\n",
				"docs/engine-seams/closed/1-x.md": "original\n",
			}
			for rel, body := range files {
				writeFile(t, filepath.Join(dir, rel), body)
			}
			git(t, dir, "init", "-q")
			git(t, dir, "add", "-A")
			git(t, dir, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "commit", "-q", "-m", "base")
			for rel, body := range tc.edits {
				writeFile(t, filepath.Join(dir, rel), body)
			}

			scriptPath := filepath.Join(t.TempDir(), "allowlist.sh")
			writeFile(t, scriptPath, script)
			cmd := exec.Command("bash", "-c", `set +e +o pipefail; source "$1"; validate_generated_only_diff`, "bash", scriptPath)
			cmd.Dir = dir
			cmd.Env = append(cleanGitEnv(), "RUNNER_TEMP="+t.TempDir())
			out, err := cmd.CombinedOutput()
			accepted := err == nil
			if _, isExit := err.(*exec.ExitError); err != nil && !isExit {
				t.Fatalf("run the allowlist check: %v", err)
			}
			if accepted != tc.accept {
				t.Errorf("allowlist accepted=%v, want %v; output:\n%s", accepted, tc.accept, out)
			}
		})
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = cleanGitEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// cleanGitEnv drops variables that would point git at the repository the
// test is running in rather than the scratch one.
func cleanGitEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GIT_DIR=") || strings.HasPrefix(kv, "GIT_WORK_TREE=") ||
			strings.HasPrefix(kv, "GIT_INDEX_FILE=") {
			continue
		}
		env = append(env, kv)
	}
	return env
}
