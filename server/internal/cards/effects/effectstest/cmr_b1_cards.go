package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmrB1WantPT asserts a creature's current power and toughness.
func CmrB1WantPT(t *testing.T, g *game.Game, id uuid.UUID, what string, power, toughness int) {
	t.Helper()
	if p, tg := PtOf(t, g, id); p != power || tg != toughness {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, tg, power, toughness)
	}
}
