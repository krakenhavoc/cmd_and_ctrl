package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerKeep answers `who`'s own-creature prompt with `keep`.
func AnswerKeep(t *testing.T, g *game.Game, who, keep uuid.UUID) {
	t.Helper()
	c := LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, who)
	if c == nil {
		t.Fatalf("%s was not asked to choose: %+v", who, g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(c.ID, who, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}
