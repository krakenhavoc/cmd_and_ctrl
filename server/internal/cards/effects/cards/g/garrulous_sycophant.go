package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garrulous Sycophant — Creature — Human Advisor {2}{B}, 1/4 (Conspiracy:
// Take the Crown, S82 older sets batch 273, #5540):
//
//	"At the beginning of your end step, if you're the monarch, each opponent
//	 loses 1 life and you gain 1 life."
//
// Faramir's end-step shape: an intervening "if" (CR 603.4) on the monarch,
// checked as the step begins and again as the ability resolves. The drain is
// the shared pair, every opponent still in the game losing 1 and then you
// gaining 1 (gained whether or not any life was lost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8301de9f-3082-400f-92d4-553536f79f49",
		Name:         "Garrulous Sycophant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondMonarch{},
				effects.AtYourEndStep("Garrulous Sycophant — each opponent loses 1 life and you gain 1 life",
					effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}))),
		},
	})
}
