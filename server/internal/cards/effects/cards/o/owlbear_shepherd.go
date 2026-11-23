package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Owlbear Shepherd — Creature — Goblin Druid {2}{G}, 1/4 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"At the beginning of your end step, if creatures you control have total
//	 power 8 or greater, draw a card."
//
// An intervening "if" (CR 603.4): checked as the end step begins and again
// as the ability resolves, over the current power of every creature its
// controller controls, Shepherd included (CR 208.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e81035b5-d922-48f8-af20-0025da4721c6",
		Name:         "Owlbear Shepherd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfTotalPowerAtLeast(8, "Owlbear Shepherd — draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
