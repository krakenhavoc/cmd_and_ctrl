package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestFrozenResourcesAreDiscountedOnlyWhenTappedAndAffected(t *testing.T) {
	me, other := seatID(0).String(), seatID(1).String()
	for _, tc := range []struct {
		name string
		card protocol.CardView
	}{
		{"mana", land(cardID(1), 0, tapped())},
		{"creature", creature(cardID(2), 0, "Bear", 2, 2, tapped())},
	} {
		t.Run(tc.name, func(t *testing.T) {
			score := func(c protocol.CardView) float64 {
				return heuristic.Score(newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(c)), me)
			}
			ordinary := score(tc.card)
			for _, restriction := range []*protocol.NoUntapView{
				{Static: true}, {Next: []string{me}},
			} {
				frozen := tc.card
				frozen.NoUntap = restriction
				if score(frozen) >= ordinary {
					t.Fatal("resource missing its next untap must be worth less than an ordinary tapped one")
				}
				frozen.Tapped = false
				upright := tc.card
				upright.Tapped = false
				if score(frozen) != score(upright) {
					t.Fatal("untapped resource is available even if its next untap will be skipped")
				}
			}
			unaffected := tc.card
			unaffected.NoUntap = &protocol.NoUntapView{Next: []string{other}}
			if score(unaffected) != ordinary {
				t.Fatal("another player's marker must not discount the controller's next untap")
			}
		})
	}
}

func TestUntapRestrictionDoesNotChangeImmediateCombatValue(t *testing.T) {
	w := heuristic.DefaultWeights()
	c := creature(cardID(3), 0, "Bear", 2, 2, tapped())
	before := w.CombatValue(&c)
	c.NoUntap = &protocol.NoUntapView{Static: true}
	if w.CombatValue(&c) != before {
		t.Fatal("future untapping must not change the creature's current combat strength")
	}
}
