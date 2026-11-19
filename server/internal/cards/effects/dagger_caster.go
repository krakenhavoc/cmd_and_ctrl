package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dagger Caster — Creature — Lizard Rogue {3}{R}, 2/3 (Ravnica: Clue
// Edition):
//
//	"When this creature enters, it deals 1 damage to each opponent and 1
//	 damage to each creature your opponents control."
//
// The creature set is read as the trigger resolves and the damage is one
// simultaneous event set (CR 120.2); a source that has left deals it as it
// last existed (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1352a81-6fd5-4c4d-ada4-b81d3a9f74c0",
		Name:         "Dagger Caster",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dagger Caster — 1 damage to each opponent and each creature your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					return damageEachMatchingAndEachOpponent(NewContext(g, item), And(Creature(), OpponentControls()), 1)
				}),
		},
	})
}
