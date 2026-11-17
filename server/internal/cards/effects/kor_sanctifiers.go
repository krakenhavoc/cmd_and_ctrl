package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Sanctifiers — Creature — Kor Cleric {2}{W}, 2/3 (Planechase, #5584):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 When this creature enters, if it was kicked, destroy target artifact
//	 or enchantment."
//
// Sun-Blessed Healer's kicked enters trigger (ThisKickedAtLeast, an
// intervening "if" read off the permanent's carried kicker record, so an
// unkicked Sanctifiers puts nothing on the stack and asks for no target)
// with Reclamation Sage's target, which is not optional here: any artifact
// or enchantment, your own included. A target that left in response is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "55d8ac5b-622c-42a9-83dc-f2c1c6c2506f",
		Name:          "Kor Sanctifiers",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Kor Sanctifiers — kicked, destroy target artifact or enchantment",
					destroyFirstLegalCardTarget),
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))),
		},
	})
}
