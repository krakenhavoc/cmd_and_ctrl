package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stdB23Mana adds mana to a pool.
func StdB23Mana(t *testing.T, g *game.Game, p *game.Player, cost string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, cost); err != nil {
		t.Fatalf("AddManaForEffect %s: %v", cost, err)
	}
}
