package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ca5d5cbf-fc4f-43dd-a2af-767f277dfef9",
		Name:            "Griffin Dreamfinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Griffin Dreamfinder — return target enchantment card from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target enchantment card from your graveyard", YouOwn(), Enchantment())),
		},
	})
}
