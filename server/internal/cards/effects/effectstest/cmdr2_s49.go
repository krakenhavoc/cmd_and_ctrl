package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func S49Library(g *game.Game, n int) {
	for _, p := range g.Seats {
		for i := 0; i < n; i++ {
			PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		}
	}
}

// s49TopCards puts one known card on top of each library: nonland for
// the seats in `nonland`, a land for the others.
func S49TopCards(g *game.Game, nonland ...int) {
	for i, p := range g.Seats {
		c := game.Card{Name: "Top Land", TypeLine: "Basic Land — Forest"}
		for _, n := range nonland {
			if n == i {
				c = game.Card{Name: "Top Spell", TypeLine: "Creature — Bear"}
			}
		}
		c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
		p.Library.PushTop(c)
	}
}

func S49Hands(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = len(p.Hand.Cards)
	}
	return out
}

func S49Attack(t *testing.T, g *game.Game, attacker uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	LockInAttacks(t, g)
	PassPriorityAroundTable(t, g)
}
