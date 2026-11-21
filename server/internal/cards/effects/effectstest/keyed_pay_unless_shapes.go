package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_pay_unless_shapes_test.go — #4264: the pay-unless shapes #4256
// left on the closure form are keyed restore points. Each test takes a
// STRICT restore (restoreRoundTrip with strict) while the prompt is open
// and answers it on the restored table; the unknown-key tests are the
// rollback case (refuseUnknownKey and its own for the any-player
// registry).

// payUnlessFor is the open pay-unless `chooser` owes, nil when none.
func PayUnlessFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
