package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "16cdba97-73cf-44ba-9d64-f27009b829f8",
		Name:          "Troublemaker Ouphe",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.G2Bargain()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.DiminisherWitchWasBargained),
					"Troublemaker Ouphe — exile target artifact or enchantment an opponent controls",
					effects.ExileFirstTarget),
				effects.TargetPermanent("target artifact or enchantment an opponent controls",
					effects.And(effects.Or(effects.Artifact(), effects.Enchantment()), effects.OpponentControls()))),
		},
	})
}
