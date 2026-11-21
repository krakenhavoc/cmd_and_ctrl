package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// upkeepPromptFor returns the open pay-unless `chooser` owes, without
// answering it — answerPayUnless answers, and these tests need the
// prompt to still be sitting there while the table tries to move.
func UpkeepPromptFor(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	t.Fatalf("no upkeep pay-unless prompt for %s", chooser)
	return nil
}
