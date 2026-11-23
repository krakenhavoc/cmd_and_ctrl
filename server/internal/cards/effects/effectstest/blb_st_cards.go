package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blbSTPickOption answers the opponent's punisher question with the
// option labelled `label`.
func BlbSTPickOption(t *testing.T, g *game.Game, chooser uuid.UUID, label string) {
	t.Helper()
	pick := LatestOptionPickFor(g, chooser)
	if pick == nil {
		t.Fatalf("%v is not asked: %+v", chooser, g.PendingChoices)
	}
	for i, o := range pick.PickOptions {
		if o.Label == label {
			AnswerOptionPick(t, g, chooser, i)
			return
		}
	}
	t.Fatalf("no option %q among %+v", label, pick.PickOptions)
}
