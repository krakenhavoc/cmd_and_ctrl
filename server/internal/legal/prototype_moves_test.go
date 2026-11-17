package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// CR 718.3 (#5656): the bot is offered a prototype card both ways,
// normal and prototyped, and the engine accepts each.
func TestEnumeratorOffersAPrototypeCardBothWays(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   "prototype-enum-oracle",
		Layout:     game.LayoutPrototype,
		Owner:      seat.ID,
		Controller: seat.ID,
		Faces: game.PrototypeFaces(game.Face{
			Name: "Enum Engine", TypeLine: "Artifact Creature — Construct", ManaCost: "{0}", Power: 4, Toughness: 4,
			OracleText: "Prototype {0} — 1/1",
		}),
	}
	c.SetFace(0)
	seat.Hand.PushTop(c)

	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, c.InstanceID)
	seen := map[int]bool{}
	for _, m := range got {
		seen[faceOf(t, m)] = true
	}
	if len(got) != 2 || !seen[0] || !seen[game.PrototypeFace] {
		t.Fatalf("prototype card offered faces %v in %d casts, want normal (0) and prototyped (1): %v", seen, len(got), labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}
