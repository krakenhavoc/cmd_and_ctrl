package coverage

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// oracle_fixture_test.go — keeps testdata/oracle_text.json true.
//
// The fixture is GENERATED: never edit it by hand. It holds the
// printed text of every catalogued oracle ID, read from the Scryfall
// dump through the same cards.Index join the server uses at boot, so
// it picks the same printing the server would. CI has no dump (it is
// ~630MB and gitignored), which is why TestAbilitiesMatchOracleText
// reads this file instead; e2e-nightly.yml has the dump and runs this
// test, so a card added without regenerating, an erratum, or a card
// that left the catalog fails there.
//
// Regenerate (from server/):
//
//	CMDCTRL_SCRYFALL_DUMP=../data/scryfall/default-cards.json \
//	  go test ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent -update-oracle
//
// Its own flag rather than the package's -update: -update also
// rewrites the census block, which CI owns (census_test.go).

var updateOracle = flag.Bool("update-oracle", false,
	"regenerate "+OracleFixturePath+" from $CMDCTRL_SCRYFALL_DUMP")

// TestOracleFixtureIsCurrent regenerates the fixture from the dump and
// fails on any difference, or rewrites it with -update-oracle.
func TestOracleFixtureIsCurrent(t *testing.T) {
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

	want, missing := buildOracleFixture(t, idx)
	if len(missing) > 0 {
		// A catalogued card the dump does not know: a mistyped oracle
		// ID, or a card newer than the dump. The first is a bug the
		// server would also hit (no image, no import); the second
		// clears when the dump refreshes.
		t.Errorf("%d catalogued oracle IDs are not in the dump:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}
	body, err := EncodeOracleFixture(want)
	if err != nil {
		t.Fatal(err)
	}

	if *updateOracle {
		if err := os.WriteFile(OracleFixturePath, body, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s: %d cards", OracleFixturePath, len(want))
		return
	}

	have, err := LoadOracleFixture(OracleFixturePath)
	if err != nil {
		t.Fatalf("%v — regenerate with -update-oracle", err)
	}
	var diffs []string
	for id, c := range want {
		h, ok := have[id]
		switch {
		case !ok:
			diffs = append(diffs, fmt.Sprintf("missing   %s (%s)", c.Name, id))
		case !sameCard(h, c):
			diffs = append(diffs, fmt.Sprintf("changed   %s (%s)", c.Name, id))
		}
	}
	for id, c := range have {
		if _, ok := want[id]; !ok {
			diffs = append(diffs, fmt.Sprintf("not in catalog %s (%s)", c.Name, id))
		}
	}
	if raw, err := os.ReadFile(OracleFixturePath); err == nil && len(diffs) == 0 && !bytes.Equal(raw, body) {
		diffs = append(diffs, "formatting differs from EncodeOracleFixture (hand-edited?)")
	}
	if len(diffs) > 0 {
		sort.Strings(diffs)
		t.Errorf(`%s disagrees with the Scryfall dump in %d places:
  %s

Regenerate it — never edit it by hand:

  CMDCTRL_SCRYFALL_DUMP=../data/scryfall/default-cards.json \
    go test ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent -update-oracle

then run TestAbilitiesMatchOracleText: an erratum can make a
registered ability wrong.`, OracleFixturePath, len(diffs), strings.Join(diffs, "\n  "))
	}
}

// buildOracleFixture reads every catalogued base oracle ID out of the
// index. Placeholder printings (art series, "Card" front cards,
// AGENTS.md §7 step 1) never win cards.Index's by-oracle join, and a
// type line of "Card" here would mean that stopped being true.
func buildOracleFixture(t *testing.T, idx *cards.Index) (map[string]OracleCard, []string) {
	t.Helper()
	out := map[string]OracleCard{}
	var missing []string
	seen := map[string]bool{}
	for _, s := range effects.All() {
		base, _ := BaseOracleID(s.OracleID)
		if seen[base] {
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
