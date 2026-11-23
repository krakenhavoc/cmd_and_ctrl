package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ActivationPrice(t *testing.T, g *game.Game, who, card uuid.UUID) string {
	t.Helper()
	price, err := g.PriceActivation(who, card, 0, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	return price.Total.String()
}
