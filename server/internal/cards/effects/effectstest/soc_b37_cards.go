package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared test helpers ---------------------------------------------------

// socB37OpenPickTarget is the open pick_target prompt, or nil.
func SocB37OpenPickTarget(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	return nil
}
