package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Griffin Dreamfinder — Creature — Griffin {3}{W}{W}, 1/4 (Born of the Gods,
// #5559):
//
//	"Flying
//	 When this creature enters, return target enchantment card from your
//	 graveyard to your hand."
//
// Corpse Cur's return, mandatory and aimed at an enchantment card: with
// no legal target the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca5d5cbf-fc4f-43dd-a2af-767f277dfef9",
		Name:            "Griffin Dreamfinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Griffin Dreamfinder — return target enchantment card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.YouOwn(), effects.Enchantment())),
		},
	})
}
