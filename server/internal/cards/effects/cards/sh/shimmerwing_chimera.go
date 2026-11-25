package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmerwing Chimera — Enchantment Creature — Chimera {3}{U}, 3/2 (Theros Beyond
// Death, #5514):
//
//	"Flying
//	 At the beginning of your upkeep, return up to one other target enchantment you
//	 control to its owner's hand."
//
// The target is announced as the trigger goes on the stack (CR 603.3d), another
// enchantment (the Chimera itself is one), and "up to one" lets the controller pick
// none; it returns to its owner's hand if it is still legal as the trigger
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "21c768fa-4f1a-4a1a-abb9-e69e17426233",
		Name:            "Shimmerwing Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Shimmerwing Chimera — return up to one other target enchantment you control to its owner's hand", effects.BounceEachLegalTarget),
				effects.Another(effects.TargetPermanent("up to one other target enchantment you control", effects.Enchantment(), effects.YouControl()).WithCount(0, 1))),
		},
	})
}
