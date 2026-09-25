package effects

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// legacy_trigger_builds_test.go — ADR 0041 P9's lint (#1497, tier 4-2).
//
// A triggered row that DECLARES its Effect is built by the engine and
// named on its stack item (game.buildTriggerItemLocked), so a table
// with it waiting on the stack is a restore point. A row that still
// computes its effect in a hand-written Build is not: the closure may
// hold a value that was only true when the ability triggered, the item
// stays an unkeyed closure, and the census counts it. So does a row
// whose TargetsFrom reads the board (TargetsFromReadsBoard), because
// restore could not rebuild its clause exactly.
//
// Every such row is listed in testdata/legacy_trigger_builds.txt, one
// per line, sorted:
//
//	<card name> | <catalog key> | <row index>
//
// The list ONLY SHRINKS, the closureClassCeilings way. The test fails
//
//   - when a listed row no longer needs listing — it declares its
//     Effect now, or it is gone: delete the line (that is the tail of
//     tier 4 working); and
//   - when a row needs listing and is not listed — a NEW hand-written
//     Build. Declare the Effect instead (AGENTS.md, "Adding a
//     triggered ability").
//
// Tail batches (tier 4-3 … 4-10) each convert a run of cards and delete
// their lines; the file is sorted by card name so parallel batches
// touch different lines. Once it is empty, NewTriggeredItem loses its
// effect parameter (4-final).
//
// Delete the stale lines with
//
//	go test ./internal/cards/effects -run TestLegacyTriggerBuildsOnlyShrink -args -update-legacy-triggers
//
// which never adds a line to an existing file.

var updateLegacyTriggers = flag.Bool("update-legacy-triggers", false,
	"delete stale lines from testdata/legacy_trigger_builds.txt (never adds one)")

const legacyTriggerBuildsFile = "testdata/legacy_trigger_builds.txt"

// productionDefKeys is every catalog key the production registry filed,
// captured by TestMain after every init() and before any test registers
// a fixture of its own.
var productionDefKeys []string

func snapshotProductionDefKeys() {
	productionDefKeys = productionDefKeys[:0]
	for k := range defs {
		productionDefKeys = append(productionDefKeys, k)
	}
	sort.Strings(productionDefKeys)
}

// needsLegacyListing is the one reading of "this row's item cannot be
// rebuilt from the row", kept in step with game.triggeredAbilityRefFor.
func needsLegacyListing(t game.TriggeredAbility) bool {
	return t.Effect == nil || t.TargetsFromReadsBoard
}

// catalogKeyNames names every production catalog key for the list: the
// card, the emblem's label, the token's name, or the card that
// declares a granted bundle.
func catalogKeyNames() map[string]string {
	names := make(map[string]string, len(registry)*2)
	for id, spec := range registry {
		names[id] = spec.Name
		if spec.Emblem != nil {
			names[game.EmblemKey(id)] = spec.Emblem.Label
		}
		for _, gr := range spec.Grants {
			names[game.GrantKey(gr.Key)] = spec.Name + " (grant " + gr.Key + ")"
		}
	}
	for slug, build := range tokenTemplatesBySlug {
		t := build()
		names[game.TokenKey(slug)] = t.Card.Name + " token"
	}
	return names
}

func legacyTriggerLine(name, key string, index int) string {
	return name + " | " + key + " | " + strconv.Itoa(index)
}

// wantLegacyTriggerLines is the set the file must equal.
func wantLegacyTriggerLines(t *testing.T) map[string]bool {
	t.Helper()
	if len(productionDefKeys) == 0 {
		t.Fatal("TestMain did not capture the production catalog keys")
	}
	names := catalogKeyNames()
	want := map[string]bool{}
	for _, key := range productionDefKeys {
		d := defs[key]
		if d == nil {
			continue
		}
		name := names[key]
		if name == "" {
			name = key
		}
		for i, tr := range d.Triggered {
			if needsLegacyListing(tr) {
				want[legacyTriggerLine(name, key, i)] = true
			}
		}
	}
	return want
}

func readLegacyTriggerLines(t *testing.T) ([]string, bool) {
	t.Helper()
	raw, err := os.ReadFile(legacyTriggerBuildsFile)
	if os.IsNotExist(err) {
		return nil, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out, true
}

const legacyTriggerBuildsHeader = `# legacy_trigger_builds.txt — ADR 0041 P9's allowlist (#1497, tier 4-2).
#
# Every catalog triggered row whose stack item cannot be rebuilt from the
# row: a hand-written Build with no declared Effect, or a TargetsFrom
# that reads the board. Such an item stays a census closure and holds a
# restore point back while it waits on the stack.
#
# One row per line, sorted: <card name> | <catalog key> | <row index>.
# The list ONLY SHRINKS. Convert a card by declaring its Effect (AGENTS.md,
# "Adding a triggered ability") and delete its line, or run
#   go test ./internal/cards/effects -run TestLegacyTriggerBuildsOnlyShrink -args -update-legacy-triggers
# which deletes stale lines and never adds one. A new hand-written Build
# fails the test instead of being added here.
`

func writeLegacyTriggerLines(t *testing.T, lines []string) {
	t.Helper()
	sort.Strings(lines)
	var b strings.Builder
	b.WriteString(legacyTriggerBuildsHeader)
	b.WriteString("\n")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	if err := os.MkdirAll(filepath.Dir(legacyTriggerBuildsFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyTriggerBuildsFile, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyTriggerBuildsOnlyShrink(t *testing.T) {
	want := wantLegacyTriggerLines(t)
	have, exists := readLegacyTriggerLines(t)

	if *updateLegacyTriggers {
		var keep []string
		if !exists {
			// Bootstrapping the list is the one time it is written
			// whole: the slice that created it.
			for l := range want {
				keep = append(keep, l)
			}
		} else {
			for _, l := range have {
				if want[l] {
					keep = append(keep, l)
				}
			}
		}
		writeLegacyTriggerLines(t, keep)
		have, exists = readLegacyTriggerLines(t)
	}
	if !exists {
		t.Fatalf("%s is missing", legacyTriggerBuildsFile)
	}

	listed := map[string]bool{}
	var problems []string
	for i, l := range have {
		if listed[l] {
			problems = append(problems, fmt.Sprintf("listed twice: %s", l))
		}
		listed[l] = true
		if i > 0 && have[i-1] > l {
			problems = append(problems, fmt.Sprintf("not sorted: %q comes after %q", l, have[i-1]))
		}
		if !want[l] {
			problems = append(problems, fmt.Sprintf("no longer needs listing — delete the line (the list only shrinks): %s", l))
		}
	}
	var missing []string
	for l := range want {
		if !listed[l] {
			missing = append(missing, l)
		}
	}
	sort.Strings(missing)
	for _, l := range missing {
		problems = append(problems, fmt.Sprintf("a triggered row computes its effect in a hand-written Build (or its TargetsFrom reads the board) and is not listed — declare TriggeredAbility.Effect instead (ADR 0041 P9): %s", l))
	}
	if len(problems) > 0 {
		t.Errorf("%s:\n  %s", legacyTriggerBuildsFile, strings.Join(problems, "\n  "))
	}
}

// TestDeclaredTriggerRowsAreStampable holds the other half: a row that
// is NOT on the list must be one the engine can name. Every production
// row that declares its Effect carries the catalog identity the
// registry stamped (fileDef), so a harvested item of it is keyed.
func TestDeclaredTriggerRowsAreStampable(t *testing.T) {
	declared, stampable := 0, 0
	for _, key := range productionDefKeys {
		d := defs[key]
		if d == nil {
			continue
		}
		for i, tr := range d.Triggered {
			if needsLegacyListing(tr) {
				continue
			}
			declared++
			if game.TriggeredAbilityRef(tr) != nil {
				stampable++
				continue
			}
			t.Errorf("%s row %d (%q) declares its Effect but the engine cannot name it", key, i, tr.Key)
		}
	}
	t.Logf("%d declared triggered rows, %d stampable", declared, stampable)
}
