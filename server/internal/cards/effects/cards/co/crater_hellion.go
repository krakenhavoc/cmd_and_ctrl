package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crater Hellion — Creature — Hellion Beast, {4}{R}{R}, 6/6:
//
//	"Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, it deals 4 damage to each other creature."
//
// "Each other creature" is every creature but this object as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d37c437-0d5c-400d-88ce-10d173b28eda",
		Name:         "Crater Hellion",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Crater Hellion", "{4}{R}{R}"),
			effects.WhenThisEnters("Crater Hellion — 4 damage to each other creature", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.NotSelf(item.SourceCardID)), 4)
			}),
		},
	})
}
