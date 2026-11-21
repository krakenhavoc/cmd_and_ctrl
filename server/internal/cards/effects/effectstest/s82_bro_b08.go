package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func BroSac(t *testing.T, g *game.Game, who, id uuid.UUID) {
	t.Helper()
	if err := g.SacrificePermanent(who, id); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
}

func GraveyardHas(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
