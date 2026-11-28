package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Cannoneers — Creature — Orc Warrior {1}{R}{R}, 1/3 (Ice Age, S82 OLD batch
// 840, #5636):
//
//	"{T}: This creature deals 2 damage to any target and 3 damage to you."
//
// Both halves are damage from the Cannoneers, dealt as the ability resolves: the
// target takes 2 (a target that is gone makes the ability do nothing, CR 608.2b)
// and the controller takes 3. A Cannoneers that has left is still the source by its
// last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ebf3eeb-1a1b-4b88-a760-887c2188d287",
		Name:         "Orcish Cannoneers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to any target and 3 damage to you.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				return g.DamageInstanceForEffect(func() error {
					for _, t := range ctx.LegalTargets() {
						if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: ctx.Controller(), Amount: 3}.Apply(ctx)
				})
			},
		}},
	})
}
