package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerMayCounter answers `who`'s "may put two counters" prompt with
// `pick` (uuid.Nil: declines).
func AnswerMayCounter(t *testing.T, g *game.Game, who, pick uuid.UUID) {
	t.Helper()
	c := LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, who)
	if c == nil {
		t.Fatalf("%s was not asked: %+v", who, g.PendingChoices)
	}
	var ids []uuid.UUID
	if pick != uuid.Nil {
		ids = []uuid.UUID{pick}
	}
	if err := g.ResolveOwnPermanents(c.ID, who, ids); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}
