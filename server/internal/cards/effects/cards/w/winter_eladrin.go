package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winter Eladrin — Creature — Faerie Elf Wizard {2}{U}, 2/2 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Gust of Wind — When this creature enters, return up to one other target
//	 creature to its owner's hand."
//
// "Other" is object identity (Another): the Eladrin never targets itself. A
// target that left or became illegal is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48ae84e7-125d-4018-afc1-cd78448d1a13",
		Name:         "Winter Eladrin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Winter Eladrin — return up to one other target creature to its owner's hand", effects.BounceTheTargetToHand),
				effects.Another(effects.TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}
