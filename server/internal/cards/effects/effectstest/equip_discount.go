package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// equipPrice is what equipping `equipment` onto `target` costs `who`.
func EquipPrice(t *testing.T, g *game.Game, who, equipment, target uuid.UUID) string {
	t.Helper()
	p, err := g.PriceActivation(who, equipment, 0, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if p.Total.ManaValue() == 0 {
		return "{0}"
	}
	return p.Total.String()
}
