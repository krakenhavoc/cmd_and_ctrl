package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerManifestDread settles whatever Oculus's trigger left: the stack
// first, then the manifest-dread pick (the first of the two looked-at
// cards), until nothing is waiting.
func AnswerManifestDread(t *testing.T, g *game.Game, who uuid.UUID) int {
	t.Helper()
	answered := 0
	for i := 0; i < 16; i++ {
		PassPriorityAroundTable(t, g)
		p := LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who)
		if p == nil {
			return answered
		}
		if err := g.ResolveChooseCards(p.ID, who, p.ChooseCards[:1]); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		answered++
	}
	t.Fatal("the manifest dread prompts never ran out")
	return answered
}
