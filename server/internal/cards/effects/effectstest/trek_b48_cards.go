package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- fixtures -------------------------------------------------------------------

// trekB48Library stacks cards on top of p's library, first named on top.
func TrekB48Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

// trekB48Restricted reports whether p carries a can't-attack grant
// protecting `protected`.
func TrekB48Restricted(p *game.Player, protected uuid.UUID) bool {
	for _, s := range p.Statics {
		if s.CantAttack.Protected == protected {
			return true
		}
	}
	return false
}

// trekB48ToAttackersOf advances to seat's declare-attackers step.
func TrekB48ToAttackersOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepDeclareAttackers {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached seat %d's declare attackers", seat)
}
