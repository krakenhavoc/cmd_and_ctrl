package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// bestow_test.go — ADR 0141 (#2862). A bestow card in hand is offered
// twice: as a creature spell with no target for its mana cost, and
// bestowed for its bestow cost with one move per creature it can
// enchant (CR 702.103b, 303.4a). dispatchAll proves CastSpell accepts
// every one.

const oracleNyxbornRollicker = "ebf2974a-963e-425a-8f8c-55f0a36984c3"

func TestEnumeratorOffersBestowOnEachCreature(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Mountain")
	bear := battlefieldCard(g, seat, creature("Bear", "{1}{G}", 2, 2))

	rollicker := handCard(seat, game.Card{
		Name: "Nyxborn Rollicker", TypeLine: "Enchantment Creature — Satyr", OracleID: oracleNyxbornRollicker,
		ManaCost: "{R}", Colors: []string{"R"}, Power: 1, Toughness: 1,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, rollicker)
	var hard, bestowed int
	sawBear := false
	for _, c := range casts {
		switch c.AlternativeCost {
		case "":
			hard++
			if len(c.Targets) != 0 {
				t.Errorf("the creature cast names targets: %+v", c)
			}
		case game.BestowKey:
			bestowed++
			if len(c.Targets) != 1 {
				t.Errorf("a bestowed cast names one creature to enchant, got %+v", c)
				continue
			}
			if c.Targets[0].ID == bear.String() {
				sawBear = true
			}
		default:
			t.Errorf("unexpected offer %+v", c)
		}
	}
	if hard != 1 || bestowed == 0 || !sawBear {
		t.Fatalf("want one creature cast and a bestowed cast onto the bear, got %+v: %v", casts, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}
