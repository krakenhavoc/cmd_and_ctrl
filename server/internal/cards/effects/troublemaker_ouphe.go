package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Troublemaker Ouphe — Creature — Ouphe {1}{G}, 2/2 (Wilds of Eldraine,
// #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 When this creature enters, if it was bargained, exile target artifact
//	 or enchantment an opponent controls."
//
// Bargain is the shared optional cost (g2Bargain), and the intervening if
// (CR 603.4) is Diminisher Witch's: an unbargained Ouphe, a reanimated one
// and a token copy put nothing on the stack and ask for no target. The
// target is Wastescape Battlemage's, exiled by ExileFirstTarget; one that
// left in response is not exiled (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "16cdba97-73cf-44ba-9d64-f27009b829f8",
		Name:          "Troublemaker Ouphe",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, diminisherWitchWasBargained),
					"Troublemaker Ouphe — exile target artifact or enchantment an opponent controls",
					ExileFirstTarget),
				TargetPermanent("target artifact or enchantment an opponent controls",
					And(Or(Artifact(), Enchantment()), OpponentControls()))),
		},
	})
}
