package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- fixtures ------------------------------------------------------------------

// trekB37Seats is the active seat and the three opponents in turn order.
func TrekB37Seats(g *game.Game) (me *game.Player, opps []*game.Player) {
	me = g.Seats[g.Turn.ActiveSeat]
	for i := 1; i < len(g.Seats); i++ {
		opps = append(opps, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	return me, opps
}

// trekB37Declare declares each attacker at its defender and locks the
// declaration in, so the attack triggers go on the stack.
func TrekB37Declare(t *testing.T, g *game.Game, attacks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for attacker, defender := range attacks {
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
}

// trekB37CastFromHand puts a card in the active seat's hand, gives it
// `lands` untapped Mountains, and casts it strictly with auto-tap from a
// main phase. Returns the card's ID and the cast error.
func TrekB37CastFromHand(t *testing.T, g *game.Game, c game.Card, lands int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Turn.Step != game.StepPostcombatMain {
		AdvanceToMain(t, g)
	}
	for i := 0; i < lands; i++ {
		PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true})
}

// trekB37Tokens is every token named `name` that p controls.
func TrekB37Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p && c.IsToken() && c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}
