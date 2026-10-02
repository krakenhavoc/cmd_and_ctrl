package lobby

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_any_color_spend_test.go — #1600. Under Chromatic
// Orrery's "you may spend mana as though it were mana of any color" the
// preview plans the cost the payment will widen — a {U}{U} with no blue
// source on the board, paid out of the Orrery's own colourless — and
// CastSpell{AutoTap} takes the same cast. The price it shows stays the
// printed one (CR 609.4b changes how a cost is paid, not the cost).
// Without the Orrery both say no.
func TestAutoTapPreviewReadsAnyColorSpend(t *testing.T) {
	const chromaticOrreryOracle = "95c3976c-33f3-490b-bfd3-7f1af2fe0416"
	for _, withOrrery := range []bool{true, false} {
		f := newPreviewFixture(t)
		f.mainPhase()
		if withOrrery {
			f.spawn(f.alice, game.ZoneBattlefield, game.Card{
				Name: "Chromatic Orrery", TypeLine: "Legendary Artifact", OracleID: chromaticOrreryOracle,
			}, 1)
		}
		f.mountains(2)
		card := f.spawn(f.alice, game.ZoneHand, game.Card{Name: "Blue Spell", TypeLine: "Sorcery", ManaCost: "{U}{U}"}, 1)[0]

		got := f.previewWith(card, "")
		if got.Cost != "{U}{U}" {
			t.Errorf("Orrery %v: preview cost %q, want the printed {U}{U}", withOrrery, got.Cost)
		}
		if withOrrery && (!got.OK || len(got.Plan) == 0) {
			t.Fatalf("under the Orrery: ok=%v plan=%v missing=%v, want a plan", got.OK, got.Plan, got.Missing)
		}
		if !withOrrery && got.OK {
			t.Fatalf("no Orrery: the preview planned %v for a {U}{U} off two Mountains", got.Plan)
		}
		if casts := f.castsWithAutoTap(card, game.CastSpellParams{}); casts != withOrrery {
			t.Errorf("Orrery %v: the preview said ok=%v and CastSpell took the cast=%v", withOrrery, got.OK, casts)
		}
	}
}
