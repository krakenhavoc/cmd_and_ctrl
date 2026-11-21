package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- helpers ---------------------------------------------------------------

// eoeQZSurveil returns player's open surveil prompt, failing the test when
// there is none or it looks at other than n cards.
func EoeQZSurveil(t *testing.T, g *game.Game, player uuid.UUID, n int) *game.PendingChoice {
	t.Helper()
	c := SurveilChoiceFor(g, player)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != n {
		t.Fatalf("surveil looked at %d cards, want %d", len(c.ScryCards), n)
	}
	return c
}

// eoeQZKeepOnTop answers a surveil prompt by keeping every card on top.
func EoeQZKeepOnTop(t *testing.T, g *game.Game, player uuid.UUID, c *game.PendingChoice) {
	t.Helper()
	if err := g.ResolveSurveil(c.ID, player, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
