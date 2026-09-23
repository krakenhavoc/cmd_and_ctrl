package lobby

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_sacrifice_test.go — #1215, the client-facing half.
// The auto-tapper refused every mana ability with a sacrifice
// component, so a Treasure was never a candidate and a board of them
// answered `ok: false` with a missing-symbols list: the cast modal
// greyed out its "Auto-tap & cast" button on a board that pays.
//
// These run through the real token catalog (effects.TreasureToken)
// rather than a hand-built shape, because "the printed Treasure comes
// through the split intact" is the claim.

// Three Treasures, no untapped lands, a {3} spell. The preview says
// yes and names all three.
func TestAutoTapPreviewPlansABoardOfTreasures(t *testing.T) {
	f := newPreviewFixture(t)
	thing := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Icy Manipulator", TypeLine: "Artifact", ManaCost: "{3}",
	}, 1)[0]
	f.spawn(f.alice, game.ZoneBattlefield, effects.TreasureToken(), 3)

	got := f.preview(thing, 0)
	if !got.OK || len(got.Plan) != 3 {
		t.Fatalf("three Treasures for {3}: ok=%v plan=%v missing=%v, want ok with three cracks",
			got.OK, got.Plan, got.Missing)
	}
}

// …and the preview shows the last-resort ordering the planner uses:
// with three Mountains beside the three Treasures, the plan is the
// Mountains. The preview is where a player sees which permanents a
// cast would spend, so the ordering being wrong here is the ordering
// being wrong where it is noticed.
func TestAutoTapPreviewPrefersLandsOverTreasures(t *testing.T) {
	f := newPreviewFixture(t)
	thing := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Icy Manipulator", TypeLine: "Artifact", ManaCost: "{3}",
	}, 1)[0]
	treasures := f.spawn(f.alice, game.ZoneBattlefield, effects.TreasureToken(), 3)
	f.mountains(3)

	got := f.preview(thing, 0)
	if !got.OK || len(got.Plan) != 3 {
		t.Fatalf("three Mountains + three Treasures for {3}: ok=%v plan=%v", got.OK, got.Plan)
	}
	for _, id := range treasures {
		for _, planned := range got.Plan {
			if planned == id.String() {
				t.Fatalf("plan %v cracks a Treasure for generic mana three Mountains could pay", got.Plan)
			}
		}
	}
}
