package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Makeshift Munitions — Enchantment {1}{R} (Commander Masters, #5948):
//
//	"{1}, Sacrifice an artifact or creature: This enchantment deals 1
//	 damage to any target."
//
// The sacrifice is a cost (CR 118.3); the damage is dealt by the
// enchantment, not by the sacrificed permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2421bec6-7647-4684-be61-d1aa951c6b4f",
		Name:         "Makeshift Munitions",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Sacrifice an artifact or creature: This enchantment deals 1 damage to any target",
			Cost: effects.Plus(effects.ManaCost("{1}"), game.AbilityCost{
				SacrificeOther: effects.SacrificeSpec("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			}),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
