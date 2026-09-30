package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agate Instigator — Creature — Lizard Rogue {1}{R}:
//
//	"Offspring {1}{R} (You may pay an additional {1}{R} as you cast
//	 this spell. If you do, when this creature enters, create a 1/1
//	 token copy of it.)
//	 Whenever another creature you control enters, this creature deals
//	 1 damage to each opponent."
//
// Both halves are shared vocabulary: the offspring keyword's cost and
// trigger, and "another creature you control enters". The token copy
// is itself an Agate Instigator, so it pings for the creatures that
// enter after it — and the original pings for the token as it enters,
// because the token is another creature you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "e1a2bd31-b800-47cd-8866-1b849a601b09",
		Name:          "Agate Instigator",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Agate Instigator"),
			WheneverAnotherCreatureEntersUnderYourControl("Agate Instigator — 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				}),
		},
	})
}
