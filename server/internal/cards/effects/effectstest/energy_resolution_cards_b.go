package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choiceOfKindFor is the first open prompt owed by `p` that is neither
// an energy prompt nor a pay_amount.
func ChoiceOfKindFor(g *game.Game, p uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == p && c.Kind != game.PendingChoicePayUnless && c.Kind != game.PendingChoicePayAmount {
			return c
		}
	}
	return nil
}
