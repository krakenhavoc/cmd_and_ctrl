package coverage

import (
	"regexp"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// flashIfLine is the printed "You may cast this spell as though it had
// flash if <condition>" (Asinine Antics, Silver Scrutiny, Timely Ward,
// Fated Clash, Bard's Company) and its inverted word order, "If <condition>,
// you may cast this spell as though it had flash" (Serpent of the Pass).
// The unconditional "as though it had flash." (Necromancy) has a full stop
// where the first has "if", and no "if" clause before it. Primal Prayers'
// "If you cast a spell this way, you may cast it as though it had flash" is
// about the spell an alternative cost casts, which is "it", not "this
// spell".
var flashIfLine = regexp.MustCompile(`(?i)\bcast this spell as though it had flash if\b|\bif\b[^.]*,\s*you may cast this spell as though it had flash\b`)

// TestNoFlashIfClauseIsAnAlternativeCost is #5207's guard over the whole
// catalog. "You may cast this spell as though it had flash if …" is a
// timing permission (CR 601.3, 601.3b–c), so a flash cast can still take
// an alternative cost from somewhere else, and only one alternative cost
// applies to a spell (CR 118.9a). Building the clause as an
// alternative-cost offer that opens the instant-speed window for its own
// claim (AlternativeCost.AsThoughFlash) makes the card weaker than
// printed, which is what #5211 marked as a caveat and #5207 retired.
//
// So every face whose printed text has the clause builds it as one of
// the two shapes that compose with any cost:
//
//   - Spec.FlashIf, the card's own conditional timing permission
//     (FlashIfYou, FlashIfXAtMost, FlashIfTargets), or
//   - an optional additional cost that opens the window when announced
//     (AdditionalCost.AsThoughFlash: FlashForMore, Quantum Reduction's
//     teamwork, OptionalBeholdOpeningFlash), for a condition on paying a
//     cost (CR 601.3c);
//
// and declares no alternative cost carrying AsThoughFlash.
func TestNoFlashIfClauseIsAnAlternativeCost(t *testing.T) {
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
		checked++
		for _, ac := range s.AlternativeCosts {
			if ac.AsThoughFlash {
				t.Errorf("%s builds \"as though it had flash if\" as the alternative cost %q: "+
					"it is a timing permission (Spec.FlashIf) or an optional cost that opens the window "+
					"(AdditionalCost.AsThoughFlash), so the flash cast can also take another alternative cost (CR 118.9a)",
					s.Name, ac.Key)
			}
		}
		opens := s.FlashIf != nil
		for _, c := range s.OptionalCosts {
			opens = opens || c.AsThoughFlash
		}
		if !opens {
			t.Errorf("%s prints \"as though it had flash if\" and builds neither a Spec.FlashIf nor an optional cost that opens the window", s.Name)
		}
	}
	if checked < 10 {
		t.Fatalf("%d catalogued faces print \"as though it had flash if\", want at least 10: "+
			"the probe proved too little (did the fixture or the clause's shape move?)", checked)
	}
	t.Logf("%d \"as though it had flash if\" faces checked", checked)
}
