package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Growing Ranks — Enchantment {2}{G/W}{G/W}:
//
//	"At the beginning of your upkeep, populate. (Create a token that's
//	 a copy of a creature token you control.)"
//
// A trigger on the stack whose effect is Populate (populate.go). With
// no creature token under your control the trigger resolves and does
// nothing (CR 701.36b); with several the controller chooses which to
// copy as the trigger resolves, not when it is put on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bec6fb31-60a3-432f-ba36-aebf172f8b27",
		Name:         "Growing Ranks",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Growing Ranks — populate", effects.Do(effects.Populate{})),
		},
	})
}
