package lobby

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_spell_spend_only_test.go — #2556, ADR 0040's
// 2026-10-07 amendment. A spell's own "spend only …" clause is read by
// the preview the way the payment reads it: Drain Life's black-only X
// and Imperiosaur's basic-lands-only mana are planned, or refused, with
// the answer CastSpell{AutoTap} gives on the same board.

const (
	drainLifeOracle   = "e75ba79f-4cc2-4ede-8641-559ab94e7e36"
	imperiosaurOracle = "e9ced5d8-8337-403f-86a3-bddb9c77d658"
)

func TestAutoTapPreviewReadsASpellsBlackOnlyX(t *testing.T) {
	for _, tc := range []struct {
		x    int
		want bool
	}{
		{2, true},
		{3, false},
	} {
		f := newPreviewFixture(t)
		f.mainPhase()
		f.lands("Swamp", 3)
		f.mountains(3)
		card := f.spawn(f.alice, game.ZoneHand, game.Card{
			Name: "Drain Life", TypeLine: "Sorcery", ManaCost: "{X}{1}{B}", OracleID: drainLifeOracle,
		}, 1)[0]

		got := f.preview(card, tc.x)
		if got.OK != tc.want {
			t.Errorf("X=%d: preview ok=%v plan=%v missing=%v, want ok=%v", tc.x, got.OK, got.Plan, got.Missing, tc.want)
		}
		params := game.CastSpellParams{XValue: tc.x, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: f.bob}}}
		if casts := f.castsWithAutoTap(card, params); casts != got.OK {
			t.Errorf("X=%d: the preview said ok=%v and CastSpell took the cast=%v", tc.x, got.OK, casts)
		}
	}
}

func TestAutoTapPreviewReadsASpellsBasicLandsOnlyMana(t *testing.T) {
	for _, withBasics := range []bool{false, true} {
		f := newPreviewFixture(t)
		f.mainPhase()
		f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: "Breeding Pool", TypeLine: "Land — Forest Island"}, 2)
		f.lands("Forest", 2)
		if withBasics {
			f.lands("Forest", 2)
		}
		card := f.spawn(f.alice, game.ZoneHand, game.Card{
			Name: "Imperiosaur", TypeLine: "Creature — Dinosaur", ManaCost: "{2}{G}{G}", OracleID: imperiosaurOracle,
		}, 1)[0]

		got := f.preview(card, 0)
		if got.OK != withBasics {
			t.Errorf("basics=%v: preview ok=%v plan=%v missing=%v", withBasics, got.OK, got.Plan, got.Missing)
		}
		if casts := f.castsWithAutoTap(card, game.CastSpellParams{}); casts != got.OK {
			t.Errorf("basics=%v: the preview said ok=%v and CastSpell took the cast=%v", withBasics, got.OK, casts)
		}
	}
}
