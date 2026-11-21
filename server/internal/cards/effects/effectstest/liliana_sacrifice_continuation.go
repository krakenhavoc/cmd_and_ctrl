package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func CommanderReturnChoicesForTest(g *game.Game) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, choice := range g.PendingChoices {
		if choice != nil && choice.Kind == game.PendingChoiceCommanderReturn {
			out = append(out, choice)
		}
	}
	return out
}
