package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chaos Maw — Creature — Hellion {5}{R}{R}, 6/6 (Hour of Devastation,
// #5534):
//
//	"When this creature enters, it deals 3 damage to each other creature."
//
// Crater Hellion's enters trigger at 3: every creature but this one, yours
// included, is dealt the damage as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04f0c489-3d1e-4ee4-a0ab-7983ae643ecb",
		Name:         "Chaos Maw",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chaos Maw — 3 damage to each other creature", func(g *game.Game, item *game.StackItem) error {
				return damageEachMatching(NewContext(g, item), And(Creature(), NotSelf(item.SourceCardID)), 3)
			}),
		},
	})
}
