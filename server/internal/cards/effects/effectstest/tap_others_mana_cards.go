package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func AssertTapped(t *testing.T, g *game.Game, want bool, ids ...uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		if got := B20Tapped(t, g, id); got != want {
			t.Errorf("card %s tapped = %v, want %v", id, got, want)
		}
	}
}
