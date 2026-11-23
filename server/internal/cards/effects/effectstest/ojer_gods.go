package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ojerOnBattlefield finds the (single) permanent of the oracle ID.
func OjerOnBattlefield(g *game.Game, oracle string) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].OracleID == oracle {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func OjerKill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
}
