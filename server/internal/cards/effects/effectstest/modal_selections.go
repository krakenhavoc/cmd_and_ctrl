package effectstest

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modal_selections_test.go — #2681 (ADR 0065 §6). A modal spell is
// offered every kind of mode selection under the default budget, on a
// board wide enough that one selection's targets alone would spend it:
// the budget goes round the selections before any gets a second
// target set, a repeatable spec offers its mixed multisets (CR
// 700.2d), and the label names the modes.

// modalTable is a main phase on seat 0's turn with ten creatures on
// the battlefield, five each for seat 0 and seat 1, no artifacts and
// nothing on the stack, and the named spell in seat 0's hand with the
// mana to cast it floating.
func ModalTable(t *testing.T, name, cost, oracle string, mana ...string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := DiscardTable(t)
	g.WithWriteLock(func() {
		for i := 0; i < 10; i++ {
			owner := me
			if i%2 == 1 {
				owner = opp
			}
			c := game.Card{
				InstanceID: uuid.New(), Name: fmt.Sprintf("Bear %d", i), TypeLine: "Creature — Bear",
				Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
				KnownBy: map[uuid.UUID]bool{},
			}
			for _, s := range g.Seats {
				c.KnownBy[s.ID] = true
			}
			g.Battlefield.PushTop(c)
		}
		for _, m := range mana {
			me.ManaPool.AddMana(game.ManaToken{Color: m})
		}
	})
	spell := HandCardOf(g, me, name, "Instant", cost, oracle)
	return g, me, spell
}
