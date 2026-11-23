package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerConfirmFor answers the open yes/no a resolving effect asked.
func AnswerConfirmFor(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == me.ID {
			if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return
		}
	}
	t.Fatalf("no yes/no prompt for %s: %+v", me.Name, g.PendingChoices)
}
