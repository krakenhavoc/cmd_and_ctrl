package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// catalogHandSpell puts a catalog spell with its printed mana cost into
// p's hand.
func CatalogHandSpell(p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func TappedCount(t *testing.T, g *game.Game, ids []uuid.UUID) int {
	t.Helper()
	n := 0
	for _, id := range ids {
		if B31Tapped(t, g, id) {
			n++
		}
	}
	return n
}
