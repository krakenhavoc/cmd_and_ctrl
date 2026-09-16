package coverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The catalog's exact-duplicate baseline. A card PR that introduces a
// new byte-identical body of six or more lines fails here, with the
// members named; the fix is to call the one that already exists (or
// to add one shared helper and call it twice), not to add the hash to
// the baseline. `go test ./internal/cards/coverage/ -update` rewrites
// the baseline from the current tree: run it when a PR REMOVES clones,
// so the file shrinks with the debt. See #584.

const (
	catalogDir    = "../effects"
	cloneBaseline = "testdata/clone_baseline.txt"
	cloneMinLines = 6
)

func TestNoNewExactClonesInTheCatalog(t *testing.T) {
	groups, err := ExactClones(catalogDir, cloneMinLines)
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.MkdirAll(filepath.Dir(cloneBaseline), 0o755); err != nil {
			t.Fatal(err)
		}
		var sb strings.Builder
		sb.WriteString("# exact-duplicate bodies in server/internal/cards/effects, one per line:\n")
		sb.WriteString("# <hash> <lines> <members> <first member>. Regenerate with -update; never add a line by hand.\n")
		for _, g := range groups {
			fmt.Fprintf(&sb, "%s %d %d %s\n", g.Hash, g.Lines, len(g.Members), g.Members[0])
		}
		if err := os.WriteFile(cloneBaseline, []byte(sb.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("baseline rewritten: %d groups", len(groups))
		return
	}
	known := map[string]bool{}
	if f, err := os.Open(cloneBaseline); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			known[strings.Fields(line)[0]] = true
		}
		_ = f.Close()
	} else {
		t.Fatalf("no baseline at %s; run `go test ./internal/cards/coverage/ -update`", cloneBaseline)
	}
	var fresh []CloneGroup
	for _, g := range groups {
		if !known[g.Hash] {
			fresh = append(fresh, g)
		}
	}
	if len(fresh) == 0 {
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d new exact-duplicate bod%s in the catalog (identical text, %d+ lines):\n\n",
		len(fresh), map[bool]string{true: "y", false: "ies"}[len(fresh) == 1], cloneMinLines)
	for _, g := range fresh {
		fmt.Fprintf(&sb, "  %d lines, %d copies:\n", g.Lines, len(g.Members))
		for _, m := range g.Members {
			fmt.Fprintf(&sb, "    %s\n", m)
		}
	}
	sb.WriteString(`
Two copies of the same body is one helper waiting to be named. Keep
one, put it in the shared vocabulary (triggers_common.go for a trigger
shape or condition, helpers.go / a mechanic-named file for a predicate
or an effect body, tokens_table.go for a token) and call it from both
places. If the duplicate is genuinely deliberate, regenerate the
baseline with -update and say why in the PR.
`)
	t.Error(sb.String())
}
