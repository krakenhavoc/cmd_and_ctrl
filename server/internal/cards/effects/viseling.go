package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viseling — Artifact Creature — Phyrexian Construct {4}, 2/2:
//
//	"At the beginning of each opponent's upkeep, this creature deals X
//	 damage to that player, where X is the number of cards in their
//	 hand minus 4."
//
// Iron Maiden on a body: the same upkeep trigger and the same shared
// damage body (damageUpkeepPlayerByHand), with the creature as the
// damage's source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05b9dccb-2364-4bca-a12e-8cb54a3f3a0b",
		Name:         "Viseling",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachOpponentsUpkeep("Viseling — deals damage to that player equal to their hand size minus 4",
				damageUpkeepPlayerByHand(handMinusFour)),
		},
	})
}
