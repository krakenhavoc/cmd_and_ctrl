package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bc891258-2f22-4248-9972-4f41905744c6",
		Name:         "Magmaw",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, Sacrifice a nonland permanent: This creature deals 1 damage to any target",
			Cost: game.AbilityCost{
				Mana:           "{1}",
				SacrificeOther: sacrificeSpec("a nonland permanent", Nonland()),
			},
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := NewContext(g, item)
				return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 1}.Apply(ctx)
			},
		}},
	})
}
