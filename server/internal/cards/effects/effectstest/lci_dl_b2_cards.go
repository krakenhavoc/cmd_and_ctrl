package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lciDL2PayUnless is the open pay_unless prompt addressed to chooser.
func LciDL2PayUnless(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, chooser)
	if c == nil {
		t.Fatal("no pay_unless prompt")
	}
	return c
}
