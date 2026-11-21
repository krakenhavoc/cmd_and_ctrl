package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Anthropede ------------------------------------------------------------------

func B24OpenPayPrompt(g *game.Game, p *game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == p.ID {
			return c
		}
	}
	return nil
}
