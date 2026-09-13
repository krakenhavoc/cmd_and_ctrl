package coverage

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_test.go is to the card-coverage roadmap what
// snapshot_drift_test.go is to the snapshot encoder: the reason the
// document will still be true in six months.
//
// The roadmap's counts were hand-typed. Nothing checked them, and
// they drifted — once by being DERIVED instead of measured (319 for a
// real 320, which is why the doc carries a "always MEASURE this line"
// warning), once by reporting registry KEYS as if they were cards
// (60 of them are MDFC back faces, so "504 specs" was 444 cards plus
// 60 land backs and the card count was overstated by ~13%).
//
// # Which way does the truth flow
//
// The document is the source of truth this test READS. It is not
// regenerated wholesale, and the test never edits a sentence a person
// wrote. What it owns is one delimited block — everything between
// BeginMarker and EndMarker — which holds only measured numbers and
// nothing else.
//
// That split is the whole design, and it is the lesson from
// internal/legal's TestTimingAgreement: `-update` is a convenience,
// and a convenience that is allowed to rewrite hand-authored claims
// converts the test into a tautology. Here `-update` can only ever
// rewrite numbers that came out of effects.All() in the first place,
// so regenerating cannot launder a wrong claim into a right-looking
// one. The prose around the block — the historical "438 → 504 at
// commit e5fc440" records, the batch table, the narrative — stays
// hand-authored and stays untouched, because it says what was true at
// a named commit and should not be quietly rewritten to today.
var update = flag.Bool("update", false,
	"rewrite the generated census block in the card-coverage roadmap")

// TestFaceSuffixMatchesCatalogKey pins this package's copy of the
// face-key convention to the game package's producer of it. The
// census splits cards from back faces by matching "#N" on the
// registry key; if game.CatalogKeyForFace ever spells the key
// differently, the census would quietly report 60 extra "cards"
// rather than failing, which is the exact error this whole file
// exists to prevent.
func TestFaceSuffixMatchesCatalogKey(t *testing.T) {
	const oracle = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"

	if got := game.CatalogKeyForFace(oracle, 0); faceSuffix.MatchString(got) {
		t.Errorf(`faceSuffix matched the FRONT-face key %q.

Face 0 keeps the bare oracle ID — that is what makes the composite
key a no-op for every single-faced card. A pattern that matches it
would count every card in the catalog as a back face.`, got)
	}
	back := game.CatalogKeyForFace(oracle, 1)
	if !faceSuffix.MatchString(back) {
		t.Errorf(`faceSuffix did not match the BACK-face key %q that
game.CatalogKeyForFace produces.

The convention moved and coverage/census.go did not. Update faceSuffix
to the new shape — until you do, TakeCensus reports every MDFC land
back as a whole card and the roadmap overstates coverage.`, back)
	}
}

// TestCensusIsInternallyConsistent is the cheap arithmetic guard: the
// two splits must both add up to the key count. It costs nothing and
// it means a later refactor of TakeCensus cannot drop a bucket
// without saying so.
func TestCensusIsInternallyConsistent(t *testing.T) {
	c := TakeCensus()
	if c.Keys == 0 {
		t.Fatal(`the registry is empty.

TakeCensus measured zero specs, which means this package is not
importing effects for real (a blank import that got dropped, a build
tag) — not that the catalog is empty.`)
	}
	if got := c.WholeCards + c.BackFaces; got != c.Keys {
		t.Errorf("whole cards (%d) + back faces (%d) = %d, want Keys = %d",
			c.WholeCards, c.BackFaces, got, c.Keys)
	}
	if got := c.Full + c.Caveats + c.Unreviewed; got != c.Keys {
		t.Errorf("full (%d) + caveats (%d) + unreviewed (%d) = %d, want Keys = %d",
			c.Full, c.Caveats, c.Unreviewed, got, c.Keys)
	}
}

// TestRoadmapCensusIsCurrent is the drift guard.
func TestRoadmapCensusIsCurrent(t *testing.T) {
	raw, err := os.ReadFile(RoadmapPath)
	if err != nil {
		t.Fatalf("read %s: %v", RoadmapPath, err)
	}
	doc := string(raw)
	want := TakeCensus().Block()

	got, ok := Extract(doc)
	if !ok {
		t.Fatalf(`%s has no generated census block.

The markers are gone, so nothing is checking the roadmap's numbers.
Paste this back into the Progress section, or run:

  go test ./internal/cards/coverage/ -update

%s`, RoadmapPath, want)
	}
	if got == want {
		return
	}

	if *update {
		next, ok := Splice(doc, want)
		if !ok {
			t.Fatalf("markers vanished between read and splice in %s", RoadmapPath)
		}
		if err := os.WriteFile(RoadmapPath, []byte(next), 0o644); err != nil {
			t.Fatalf("write %s: %v", RoadmapPath, err)
		}
		t.Logf("updated the census block in %s", RoadmapPath)
		return
	}

	c := TakeCensus()
	t.Errorf(`the card-coverage roadmap disagrees with the live catalog.

%s is out of date. The catalog moves under it every time a card PR
merges, and the numbers in that file are read by people deciding what
to work on next.

Measured right now, from len(effects.All()):

  registry keys          %d
  whole cards            %d   <- the number to quote as "cards we automate"
  MDFC back faces        %d   <- half a card each; those cards are still gaps
  declared full          %d
  declared caveats       %d
  declared unreviewed    %d

Fix it mechanically:

  go test ./internal/cards/coverage/ -update

That rewrites ONLY the block between the two HTML comment markers.
Every hand-written sentence in the roadmap — including the historical
"N → M at commit <sha>" records, which are true of their commit and
must not be retconned — is left exactly as it is.

%s`,
		RoadmapPath,
		c.Keys, c.WholeCards, c.BackFaces, c.Full, c.Caveats, c.Unreviewed,
		diff(got, want))
}

// diff renders the two blocks line by line, marking the first line
// that differs. A whole-block dump is unreadable at twelve lines of
// markdown; the reader wants the number that moved.
func diff(got, want string) string {
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	var b strings.Builder
	b.WriteString("Line-by-line (- in the doc, + measured):\n")
	n := len(g)
	if len(w) > n {
		n = len(w)
	}
	for i := 0; i < n; i++ {
		var gl, wl string
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl == wl {
			continue
		}
		if gl != "" {
			b.WriteString("  - " + gl + "\n")
		}
		if wl != "" {
			b.WriteString("  + " + wl + "\n")
		}
	}
	return b.String()
}
