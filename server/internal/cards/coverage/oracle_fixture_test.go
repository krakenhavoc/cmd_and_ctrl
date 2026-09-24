package coverage

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// oracle_fixture_test.go — keeps testdata/oracle/ true.
//
// The fixture is GENERATED: never edit it by hand. It holds one file
// per catalogued base oracle ID, testdata/oracle/<oracle_id>.json, with
// that card's printed text, read from the Scryfall dump through the
// same cards.Index join the server uses at boot, so it picks the same
// printing the server would. CI has no dump (it is ~630MB and
// gitignored), which is why TestAbilitiesMatchOracleText reads these
// files instead; e2e-nightly.yml has the dump and runs this test, so a
// card added without its file, an erratum, or a file for a card that
// left the catalog fails there.
//
// Every catalogued card needs its own file, and a PR adds only its own
// cards' files (#1542). Regenerate just those (from server/):
//
//	CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json \
//	  go test ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent \
//	  -update-oracle -oracle-ids=<id>,<id>
//
// Drop -oracle-ids to regenerate every file; that also deletes the
// file of every card no longer catalogued.
//
// Its own flag rather than the package's -update: -update also
// rewrites the census block, which CI owns (census_test.go).

var (
	updateOracle = flag.Bool("update-oracle", false,
		"regenerate the per-card files under "+OracleFixtureDir+" from $CMDCTRL_SCRYFALL_DUMP")
	oracleIDsFlag = flag.String("oracle-ids", "",
		"comma-separated base oracle IDs: TestOracleFixtureIsCurrent checks (and with -update-oracle writes) only these cards' files")
)

// oracleRegenCommand is the command the failure messages print, run
// from server/. The dump path is spelled through $PWD because the test
// binary runs in the package directory, where a relative
// "../data/…" would not resolve. With ids it regenerates only those
// cards' files, which is what a card PR wants.
func oracleRegenCommand(ids []string) string {
	cmd := "  CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json \\\n" +
		"    go test ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent -update-oracle"
	if len(ids) > 0 {
		cmd += " \\\n    -oracle-ids=" + strings.Join(ids, ",")
	}
	return cmd
}

// parseOracleIDs reads -oracle-ids into a set of base oracle IDs, nil
// when the flag is empty (every card). A "#<face>" suffix is dropped,
// so a Spec key can be pasted as-is.
func parseOracleIDs(s string) (map[string]bool, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	out := map[string]bool{}
	for _, part := range strings.Split(s, ",") {
		base, _ := BaseOracleID(strings.ToLower(strings.TrimSpace(part)))
		if base == "" {
			continue
		}
		if !isOracleID(base) {
			return nil, fmt.Errorf("-oracle-ids: %q is not an oracle ID", part)
		}
		out[base] = true
	}
	return out, nil
}

// TestOracleFixtureIsCurrent regenerates the fixture from the dump and
// fails on any difference, or rewrites it with -update-oracle.
// -oracle-ids narrows both to the named cards.
func TestOracleFixtureIsCurrent(t *testing.T) {
	scope, err := parseOracleIDs(*oracleIDsFlag)
	if err != nil {
		t.Fatal(err)
	}
	dump := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if dump == "" {
		if *updateOracle {
			t.Fatal("-update-oracle needs CMDCTRL_SCRYFALL_DUMP pointing at the Scryfall default-cards dump")
		}
		t.Skip("CMDCTRL_SCRYFALL_DUMP unset; the fixture is checked against the dump nightly")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(dump); err != nil {
		t.Fatalf("load %s: %v", dump, err)
	}

	catalogued := cataloguedOracleIDs(effects.All())
	want, missing := buildOracleFixture(t, idx, scope)
	if len(missing) > 0 {
		// A catalogued card the dump does not know: a mistyped oracle
		// ID, or a card newer than the dump. The first is a bug the
		// server would also hit (no image, no import); the second
		// clears when the dump refreshes. Its file, if any, is left
		// alone either way.
		t.Errorf("%d catalogued oracle IDs are not in the dump:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}

	if *updateOracle {
		wrote, removed, err := writeOracleFixture(OracleFixtureDir, want, catalogued, scope)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range wrote {
			t.Logf("wrote   %s (%s)", OracleFixtureFile(id), want[id].Name)
		}
		for _, id := range removed {
			t.Logf("removed %s (no longer catalogued)", OracleFixtureFile(id))
		}
		t.Logf("%s: %d cards in scope, %d files written, %d removed", OracleFixtureDir, len(want), len(wrote), len(removed))
		return
	}

	have, raw, err := LoadOracleFixtureFS(os.DirFS(OracleFixtureDir), ".")
	if err != nil {
		t.Fatalf("%v\n\nregenerate the fixture:\n\n%s", err, oracleRegenCommand(nil))
	}
	diffs, ids := diffOracleFixture(want, have, raw, catalogued, scope)
	if len(diffs) > 0 {
		t.Errorf(`%s disagrees with the Scryfall dump in %d places:
  %s

Regenerate — never edit a file by hand. Just these cards:

%s

or every card (also deletes files for cards that left the catalog):

%s

then run TestAbilitiesMatchOracleText: an erratum can make a
registered ability wrong.`, OracleFixtureDir, len(diffs), strings.Join(diffs, "\n  "),
			oracleRegenCommand(ids), oracleRegenCommand(nil))
	}
}

// diffOracleFixture compares the generated fixture (want) with the
// files on disk (have, raw), within scope (nil: every card). It
// returns one line per disagreement and the IDs whose files a scoped
// regeneration would fix.
//
// A file for an ID that is not catalogued is reported as "not in
// catalog" — the stale file a card that left the catalog leaves
// behind. That is the one disagreement this test can see that
// TestOracleFixtureCoversRegistry cannot.
func diffOracleFixture(want, have map[string]OracleCard, raw map[string][]byte,
	catalogued, scope map[string]bool) (diffs, ids []string) {
	inScope := func(id string) bool { return scope == nil || scope[id] }
	for id, c := range want {
		h, ok := have[id]
		switch {
		case !ok:
			diffs = append(diffs, fmt.Sprintf("missing   %s (%s)", c.Name, id))
		case !sameCard(h, c):
			diffs = append(diffs, fmt.Sprintf("changed   %s (%s)", c.Name, id))
		default:
			if enc, err := EncodeOracleCard(c); err == nil && bytes.Equal(enc, raw[id]) {
				continue
			}
			diffs = append(diffs, fmt.Sprintf("formatting %s (%s): not what EncodeOracleCard writes (hand-edited?)", c.Name, id))
		}
		ids = append(ids, id)
	}
	for id, c := range have {
		if inScope(id) && !catalogued[id] {
			diffs = append(diffs, fmt.Sprintf("not in catalog %s (%s)", c.Name, id))
			ids = append(ids, id)
		}
	}
	sort.Strings(diffs)
	sort.Strings(ids)
	return diffs, ids
}

// writeOracleFixture brings dir in line with want, touching only the
// files in scope (nil: every file). It writes each card of want whose
// file is missing or differs, and removes the file of each in-scope ID
// that is not catalogued. It never removes the file of a catalogued
// card, even one the dump did not know: that is a stale dump, and
// deleting the file would turn it into a TestOracleFixtureCoversRegistry
// failure. Returns the IDs written and removed, sorted.
func writeOracleFixture(dir string, want map[string]OracleCard, catalogued, scope map[string]bool) (wrote, removed []string, err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	for id, c := range want {
		if scope != nil && !scope[id] {
			continue
		}
		body, err := EncodeOracleCard(c)
		if err != nil {
			return nil, nil, err
		}
		p := filepath.Join(dir, id+".json")
		if old, err := os.ReadFile(p); err == nil && bytes.Equal(old, body) {
			continue
		}
		if err := os.WriteFile(p, body, 0o644); err != nil {
			return nil, nil, err
		}
		wrote = append(wrote, id)
	}
	candidates := map[string]bool{}
	if scope != nil {
		candidates = scope
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, nil, err
		}
		for _, e := range entries {
			if id, ok := strings.CutSuffix(e.Name(), ".json"); ok && isOracleID(id) {
				candidates[id] = true
			}
		}
	}
	for id := range candidates {
		if catalogued[id] {
			continue
		}
		err := os.Remove(filepath.Join(dir, id+".json"))
		switch {
		case err == nil:
			removed = append(removed, id)
		case !errors.Is(err, fs.ErrNotExist):
			return nil, nil, err
		}
	}
	sort.Strings(wrote)
	sort.Strings(removed)
	return wrote, removed, nil
}

// TestOracleFixtureCoversRegistry is the cheap half of
// TestOracleFixtureIsCurrent: it runs in ordinary PR CI, with no
// Scryfall dump, and fails when a registered catalog oracle ID has no
// file in testdata/oracle/ at all. It can't tell a STALE file from a
// fresh one — that needs the dump, which is what
// TestOracleFixtureIsCurrent is for — but a MISSING file is visible
// from the registry alone, and a missing row is exactly the shape
// #1477 and #1496 both were: a card merged, nobody regenerated the
// fixture, and the gap sat invisible until the nightly dump run caught
// it.
//
// Its notion of "should have a file" mirrors buildOracleFixture's
// selection exactly (dedup by base oracle ID via BaseOracleID, one
// file per distinct base — the generator writes one for every
// catalogued card, not only ones with activated or loyalty abilities),
// so this test's pass/fail agrees with what -update-oracle would write.
func TestOracleFixtureCoversRegistry(t *testing.T) {
	have, err := LoadOracleFixture(OracleFixtureDir)
	if err != nil {
		t.Fatalf("%v\n\nregenerate the fixture (needs the dump):\n\n%s", err, oracleRegenCommand(nil))
	}
	missing, ids := missingOracleFiles(effects.All(), have)
	if len(missing) == 0 {
		return
	}
	t.Errorf(`%d catalogued oracle IDs have no file in %s:
  %s

Every catalogued card needs its own file. Generate exactly these (needs
the Scryfall dump; never write one by hand):

%s`, len(missing), OracleFixtureDir, strings.Join(missing, "\n  "), oracleRegenCommand(ids))
}

// missingOracleFiles lists the catalogued base oracle IDs with no
// fixture entry, as "<name> (<file>)" lines and as bare IDs, both
// sorted.
func missingOracleFiles(specs []effects.Spec, have map[string]OracleCard) (lines, ids []string) {
	seen := map[string]bool{}
	for _, s := range specs {
		base, _ := BaseOracleID(s.OracleID)
		if seen[base] {
			continue
		}
		seen[base] = true
		if _, ok := have[base]; !ok {
			lines = append(lines, fmt.Sprintf("%s (%s)", s.Name, OracleFixtureFile(base)))
			ids = append(ids, base)
		}
	}
	sort.Strings(lines)
	sort.Strings(ids)
	return lines, ids
}

// cataloguedOracleIDs is the set of distinct base oracle IDs the
// registry declares — the set that should have files.
func cataloguedOracleIDs(specs []effects.Spec) map[string]bool {
	out := map[string]bool{}
	for _, s := range specs {
		base, _ := BaseOracleID(s.OracleID)
		out[base] = true
	}
	return out
}

// buildOracleFixture reads every catalogued base oracle ID in scope
// (nil: all) out of the index. Placeholder printings (art series,
// "Card" front cards, AGENTS.md §7 step 1) never win cards.Index's
// by-oracle join, and a type line of "Card" here would mean that
// stopped being true.
func buildOracleFixture(t *testing.T, idx *cards.Index, scope map[string]bool) (map[string]OracleCard, []string) {
	t.Helper()
	out := map[string]OracleCard{}
	var missing []string
	seen := map[string]bool{}
	for _, s := range effects.All() {
		base, _ := BaseOracleID(s.OracleID)
		if seen[base] || (scope != nil && !scope[base]) {
			continue
		}
		seen[base] = true
		id, err := uuid.Parse(base)
		if err != nil {
			missing = append(missing, fmt.Sprintf("%s (%s): not a UUID", s.Name, base))
			continue
		}
		c, ok := idx.FindByOracleID(id)
		if !ok {
			missing = append(missing, fmt.Sprintf("%s (%s)", s.Name, base))
			continue
		}
		if c.TypeLine == "Card" || c.TypeLine == "Card // Card" || c.Layout == "art_series" {
			t.Errorf("%s (%s): the by-oracle join picked a placeholder printing (%s, %q)", s.Name, base, c.Layout, c.TypeLine)
			continue
		}
		oc := OracleCard{Name: c.Name}
		if len(c.CardFaces) > 0 {
			for _, f := range c.CardFaces {
				oc.Faces = append(oc.Faces, OracleFace{Name: f.Name, Text: f.OracleText})
			}
		} else {
			oc.Text = c.OracleText
		}
		out[base] = oc
	}
	sort.Strings(missing)
	return out, missing
}

func sameCard(a, b OracleCard) bool {
	if a.Name != b.Name || a.Text != b.Text || len(a.Faces) != len(b.Faces) {
		return false
	}
	for i := range a.Faces {
		if a.Faces[i] != b.Faces[i] {
			return false
		}
	}
	return true
}
