package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brooding Saurian — Creature — Lizard {2}{G}{G}, 4/4 (Commander 2013, S82
// older sets):
//
//	"At the beginning of each end step, each player gains control of all
//	 nontoken permanents they own."
//
// Every nontoken permanent whose owner doesn't control it goes back to its
// owner, indefinitely, as one resolution at each end step (CR 611.2a). A
// token, and a permanent already under its owner's control, are left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fb7f844-edaf-43ef-9121-318baf9ec9ce",
		Name:         "Brooding Saurian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Brooding Saurian — each player gains control of all nontoken permanents they own",
				effects.OldB384ReturnOwnedNontokenPermanents),
		},
	})
}
