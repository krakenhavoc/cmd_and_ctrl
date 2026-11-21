package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerPermanentPicks answers every open creature-pick leg with the
// creatures `pick` names for the leg's board (none for the rest).
func AnswerPermanentPicks(t *testing.T, g *game.Game, chooser uuid.UUID, pick map[uuid.UUID][]uuid.UUID) int {
	t.Helper()
	legs := 0
	for i := 0; i < 8; i++ {
		var c *game.PendingChoice
		g.ReadSnapshot(func() {
			for _, pc := range g.PendingChoices {
				if pc != nil && (pc.Kind == game.PendingChoiceOwnPermanents || pc.Kind == game.PendingChoiceTheirPermanents) {
					c = pc
					return
				}
			}
		})
		if c == nil {
			return legs
		}
		legs++
		picks := pick[c.FromPlayer]
		var err error
		if c.Kind == game.PendingChoiceOwnPermanents {
			err = g.ResolveOwnPermanents(c.ID, chooser, picks)
		} else {
			err = g.ResolveTheirPermanents(c.ID, chooser, picks)
		}
		if err != nil {
			t.Fatalf("answering the leg over %s: %v", c.FromPlayer, err)
		}
	}
	t.Fatal("permanent-pick prompts did not stop")
	return legs
}
