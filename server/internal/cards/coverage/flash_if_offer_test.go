package coverage

import (
	"regexp"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// flashIfLine is the printed "You may cast this spell as though it had
// flash if <condition>" (Asinine Antics, Silver Scrutiny, Timely Ward,
// Fated Clash, Bard's Company). The unconditional "as though it had
// flash." (Necromancy) has a full stop where this has "if".
var flashIfLine = regexp.MustCompile(`(?i)\bcast this spell as though it had flash if\b`)

// TestEveryFlashIfOfferCardDeclaresItsCaveat is #5207's probe over the
// whole catalog. "You may cast this spell as though it had flash if …"
// is a timing permission (CR 601.3b–c), so on the printed card a flash
// cast can still take an alternative cost from somewhere else. The
// catalog builds it as an alternative-cost offer that opens the
// instant-speed window for its own claim (AlternativeCost.AsThoughFlash:
// FlashForMore, FlashIfXAtMost, FlashIfTargets, a Condition-gated
// offer), and only one alternative cost applies to a spell (CR 118.9a).
// So every face whose printed text has the clause and whose Spec builds
// it that way declares CompletenessCaveats with a caveat that says so.
//
// A card that builds the clause another way — an optional additional
// cost that opens the window (AdditionalCost.AsThoughFlash, Quantum
// Reduction's teamwork) — has no alternative-cost offer that opens it,
// and this probe leaves it alone.
func TestEveryFlashIfOfferCardDeclaresItsCaveat(t *testing.T) {
	fx := loadOracleFixture(t)
	checked := 0
	for _, s := range effects.All() {
		base, face := BaseOracleID(s.OracleID)
		oc, ok := fx[base]
		if !ok {
			continue
		}
		faces := oc.AllFaces()
		if face >= len(faces) || !flashIfLine.MatchString(faces[face].Text) {
			continue
		}
		offers := false
		for _, ac := range s.AlternativeCosts {
			offers = offers || ac.AsThoughFlash
		}
		if !offers {
			continue
		}
		checked++
		if s.Completeness != effects.CompletenessCaveats {
			t.Errorf("%s builds \"as though it had flash if\" as an alternative-cost offer and is %v: "+
				"declare CompletenessCaveats (the flash cast can't take another alternative cost, CR 118.9a)",
				s.Name, s.Completeness)
			continue
		}
		said := false
		for _, cv := range s.Caveats {
			said = said || strings.Contains(cv, "another alternative cost")
		}
		if !said {
			t.Errorf("%s: no caveat says the flash cast can't be combined with another alternative cost: %q",
				s.Name, s.Caveats)
		}
	}
	if checked < 6 {
		t.Fatalf("%d catalogued faces build \"as though it had flash if\" as an offer, want at least 6: "+
			"the probe proved too little (did the fixture or the clause's shape move?)", checked)
	}
	t.Logf("%d \"as though it had flash if\" offer faces checked", checked)
}
