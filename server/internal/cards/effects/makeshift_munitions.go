package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2421bec6-7647-4684-be61-d1aa951c6b4f",
		Name:         "Makeshift Munitions",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, Sacrifice an artifact or creature: This enchantment deals 1 damage to any target",
			Cost: Plus(ManaCost("{1}"), game.AbilityCost{
				SacrificeOther: sacrificeSpec("an artifact or creature", Or(Artifact(), Creature())),
			}),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
