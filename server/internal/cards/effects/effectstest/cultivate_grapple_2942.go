package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cultivate_grapple_2942.go — Cultivate-style search test helpers shared across test packages.

// AnswerOrDecline answers the open search prompt with forests[idx], or
// takes nothing for idx < 0.
func AnswerOrDecline(t *testing.T, g *game.Game, chooser uuid.UUID, idx int, forests []uuid.UUID) {
	t.Helper()
	if idx < 0 {
		AnswerSearchFailToFind(t, g, chooser)
		return
	}
	AnswerSearchByID(t, g, chooser, forests[idx])
}
