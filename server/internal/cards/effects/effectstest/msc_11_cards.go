package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc11Cast casts a spell with the given cost and colours from the active
// seat's hand, leaving it on the stack.
func Msc11Cast(t *testing.T, g *game.Game, name, typeLine, manaCost string, colors []string) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := HandCardFull(me, name, typeLine, manaCost, "", colors)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
