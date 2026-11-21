package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gandalf, Spark Starter — Legendary Creature — Avatar Wizard {4}{R}{R}, 4/3
// (The Hobbit, HOB):
//
//	"Reach
//	 When Gandalf enters, he deals 3 damage divided as you choose among
//	 one, two, or three targets."
//
// Inferno Titan's enters half: one to three any-target clauses with the
// three damage divided at announce (CR 601.2d, 700.2i), so each chosen
// target takes at least 1, and the shares are read back as the trigger
// resolves (DealDividedDamage). Gandalf deals the damage. Reach is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd1a9d87-4c91-4456-8d88-19e0d6e0b200",
		Name:            "Gandalf, Spark Starter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Gandalf, Spark Starter — 3 damage divided among one, two, or three targets",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DealDividedDamage(effects.NewContext(g, item))
				}),
				effects.TargetAny().WithCount(1, 3).Dividing(effects.Divide(3))),
		},
	})
}
