package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmtB10AnswerAttackTarget answers "what does it attack?" with the first
// option when the put asked it.
func TmtB10AnswerAttackTarget(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	if LatestOptionPickFor(g, chooser) != nil {
		AnswerOptionPick(t, g, chooser, 0)
	}
}
