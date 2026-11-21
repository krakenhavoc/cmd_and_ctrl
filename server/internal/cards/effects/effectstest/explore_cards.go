package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerExploreQuestion answers the open "put it into your graveyard?"
// prompt.
func AnswerExploreQuestion(t *testing.T, g *game.Game, chooser uuid.UUID, accept bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			if err := g.ResolveConfirm(c.ID, chooser, accept); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return
		}
	}
	t.Fatal("no explore question is open")
}

func PutOnTopForTest(p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Library.PushTop(c)
	return c.InstanceID
}
