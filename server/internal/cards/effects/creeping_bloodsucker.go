package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creeping Bloodsucker — Creature — Vampire {1}{B}, 1/2 (EDHREC rank
// 1537):
//
//	"At the beginning of your upkeep, this creature deals 1 damage to
//	 each opponent. You gain life equal to the damage dealt this way."
//
// A drain on a two-drop. The damage is DAMAGE from the creature, not
// life loss — a prevention shield stops it and a damage doubler grows
// it — and "the damage dealt this way" is read back as the sum of
// each opponent's actual life change, so what the Bloodsucker's
// controller gains is what actually landed, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bdddaf9-579a-4588-b5b3-faa189d4bdcc",
		Name:         "Creeping Bloodsucker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Creeping Bloodsucker — 1 damage to each opponent, gain that much life", func(g *game.Game, item *game.StackItem) error {
				return b14DamageEachOpponentGainThatMuch(g, item, 1)
			}),
		},
	})
}
