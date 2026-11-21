package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blightPrompt is the open blight pick addressed to chooser, or nil.
func BlightPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	c := LatestChooseCardsFor(g, chooser)
	if c == nil || c.ChooseDestination != game.ChooseBlight {
		return nil
	}
	return c
}

func MinusCounters(g *game.Game, id uuid.UUID) int {
	c := FindBattlefieldCardByID(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterMinusOne]
}
