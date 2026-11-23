package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boltbender — Creature — Goblin Wizard {3}{R}, 4/2 (Murders at Karlov
// Manor Commander, #5368):
//
//	"Disguise {1}{R}
//	 When this creature is turned face up, you may choose new targets for
//	 any number of other spells and/or abilities."
//
// Disguise is the engine's (CR 702.168a). The trigger offers an optional
// CR 115.7c choose-new-targets prompt for each other spell or ability on
// the stack that has a target; the controller answers or declines each.
// An item with nowhere else to point is skipped without a prompt
// (CR 115.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "055e1799-8188-48f4-854e-d3e28363084f",
		Name:             "Boltbender",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Boltbender — you may choose new targets for any number of other spells and/or abilities",
				effects.Cmdr2S38ChooseNewTargetsForOthers),
		},
	})
}
