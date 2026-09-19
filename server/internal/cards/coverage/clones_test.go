package coverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
//
// The baseline records the hash, the body's length, how many copies
// there are and the FILES they are declared in — no line numbers
// (#895). The human-readable report below still prints
// "file.go:closure@line", because a reader wants to go straight to
// the body; the file does not, because a line number makes every
// unrelated edit above a listed closure a diff.

const (
	catalogDir    = "../effects"
	cloneBaseline = "testdata/clone_baseline.txt"
	cloneMinLines = 6
)

// groupIsFresh reports whether g is not accounted for by the
// baseline: a hash the baseline has never recorded, or one it has
// recorded with fewer members than g now has. A group that shrank, or
// held steady, is not fresh — #895's -update rewrites a smaller
// baseline for that, it does not fail the gate.
//
// #786: the baseline's member count used to be written but never
// read back, so a third, fourth or tenth copy of an already-known
// body passed silently. known maps a baseline hash to that recorded
// count.
func groupIsFresh(known map[string]int, g CloneGroup) bool {
	n, ok := known[g.Hash]
	if !ok {
		return true
	}
	return len(g.Members) > n
}

func TestNoNewExactClonesInTheCatalog(t *testing.T) {
	groups, err := ExactClones(catalogDir, cloneMinLines)
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.MkdirAll(filepath.Dir(cloneBaseline), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cloneBaseline, []byte(FormatCloneBaseline(groups)), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("baseline rewritten: %d groups", len(groups))
		return
	}
	known := map[string]int{}
	if f, err := os.Open(cloneBaseline); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			n, err := strconv.Atoi(fields[2])
			if err != nil {
				continue
			}
			known[fields[0]] = n
		}
		_ = f.Close()
	} else {
		t.Fatalf("no baseline at %s; run `go test ./internal/cards/coverage/ -update`", cloneBaseline)
	}
	var fresh []CloneGroup
	for _, g := range groups {
		if groupIsFresh(known, g) {
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

// #786: a baseline group that already knows a hash must still catch a
// grown group — a third, fourth or tenth copy of a body the baseline
// already lists. Before this, the gate kept only the hash and never
// looked at the recorded member count, so growth passed silently.
func TestGroupIsFreshComparesMemberCounts(t *testing.T) {
	known := map[string]int{"h": 3}
	tests := []struct {
		name    string
		members int
		want    bool
	}{
		{"grown from 3 to 4 members fails", 4, true},
		{"steady at 3 members passes", 3, false},
		{"shrunk to 2 members passes", 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			members := make([]string, tt.members)
			for i := range members {
				members[i] = fmt.Sprintf("m%d.go:func F", i)
			}
			g := CloneGroup{Hash: "h", Lines: 6, Members: members}
			if got := groupIsFresh(known, g); got != tt.want {
				t.Errorf("groupIsFresh(%d members against baseline %d) = %v, want %v",
					tt.members, known["h"], got, tt.want)
			}
		})
	}
}

// A hash the baseline has never seen is fresh, same as before #786.
func TestGroupIsFreshNewHashFails(t *testing.T) {
	known := map[string]int{}
	g := CloneGroup{Hash: "new", Lines: 6, Members: []string{"a.go:func F", "b.go:func G"}}
	if !groupIsFresh(known, g) {
		t.Error("a hash absent from the baseline should be fresh")
	}
}

// #895: the baseline does not move when a listed body does.
//
// The failure this replaces: the location column recorded
// "file.go:closure@214", so an edit ANYWHERE above that closure — a
// comment, a new helper, a reordered import — shifted the number and
// `-update` rewrote the file. The gate never read it; it keys on the
// hash. Three agents in one week hit the noise (#636, then #870/#877
// in PR #878).
//
// The two synthetic sets below are the SAME clone groups reported from
// two trees that differ only in where the bodies sit. A baseline keyed
// on lines renders them differently; one keyed on the hash and the
// declaring files renders them identically, which is the whole fix —
// and this is what would catch a regression that put a line number
// back.
func TestCloneBaselineIsInsensitiveToLineNumbers(t *testing.T) {
	before := []CloneGroup{
		{Hash: "00d76bde97807c32", Lines: 6, Members: []string{
			"haywire_mite.go:closure@34", "tattered_mummy.go:closure@21",
		}},
		{Hash: "19705d5be8187b11", Lines: 6, Members: []string{
			"arbor_elf.go:closure@28", "arbor_elf.go:func plainMana", "utopia_sprawl.go:closure@11",
		}},
	}
	// The same bodies, further down each file, and reported in a
	// different order within the second group.
	after := []CloneGroup{
		{Hash: "00d76bde97807c32", Lines: 6, Members: []string{
			"haywire_mite.go:closure@91", "tattered_mummy.go:closure@140",
		}},
		{Hash: "19705d5be8187b11", Lines: 6, Members: []string{
			"arbor_elf.go:func plainMana", "utopia_sprawl.go:closure@300", "arbor_elf.go:closure@402",
		}},
	}

	got, want := FormatCloneBaseline(after), FormatCloneBaseline(before)
	if got != want {
		t.Errorf("moving the bodies changed the baseline:\n--- before\n%s--- after\n%s", want, got)
	}
	if strings.Contains(got, "@") {
		t.Errorf("the baseline carries a line number:\n%s", got)
	}
	// A group's location column is its deduplicated, sorted file set.
	if !strings.Contains(got, "19705d5be8187b11 6 3 arbor_elf.go,utopia_sprawl.go\n") {
		t.Errorf("the files column is not the sorted, deduplicated set:\n%s", got)
	}
}

// The gate itself is unchanged: #895 changed what the baseline
// RECORDS, not what it catches. Every line still carries a hash the
// reader keys on, in the four columns the writer emits, and a hash no
// body could have is not in it — so a new group is still fresh, and
// fresh is still a failure.
func TestCloneBaselineFileKeepsTheGatesKey(t *testing.T) {
	f, err := os.Open(cloneBaseline)
	if err != nil {
		t.Fatalf("no baseline at %s: %v", cloneBaseline, err)
	}
	defer func() { _ = f.Close() }()

	known := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 4 {
			t.Errorf("baseline line %q has %d fields, want hash, lines, copies, files", line, len(fields))
			continue
		}
		if strings.Contains(fields[3], "@") {
			t.Errorf("baseline line %q records a line number", line)
		}
		known[fields[0]] = true
	}
	if len(known) == 0 {
		t.Fatal("the baseline recorded no groups")
	}
	fresh := CloneGroup{Hash: "0000000000000000", Lines: 9, Members: []string{"a.go:func x", "b.go:func y"}}
	if known[fresh.Hash] {
		t.Fatal("a hash no body could have is in the baseline, so a new group would pass the gate")
	}
}
