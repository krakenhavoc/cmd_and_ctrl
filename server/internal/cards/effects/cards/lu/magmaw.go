package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magmaw — Creature — Elemental {3}{R}{R}, 4/4:
//
//	"{1}, Sacrifice a nonland permanent: This creature deals 1 damage to
//	 any target."
//
// Goblin Bombardment's ping with a mana payment and a nonland permanent
// (itself included) as the sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc891258-2f22-4248-9972-4f41905744c6",
		Name:         "Magmaw",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Sacrifice a nonland permanent: This creature deals 1 damage to any target",
			Cost: game.AbilityCost{
				Mana:           "{1}",
				SacrificeOther: effects.SacrificeSpec("a nonland permanent", effects.Nonland()),
			},
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 1}.Apply(ctx)
			},
		}},
	})
}
