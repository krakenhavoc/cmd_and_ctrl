package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Bombardment — Enchantment for {1}{R}:
//
//	"Sacrifice a creature: This enchantment deals 1 damage to any
//	target."
//
// The S21 exit-criteria card and the canonical free sacrifice
// outlet. No tap, no mana — the whole cost is the creature, so it
// can be activated any number of times at instant speed, which is
// what makes the aristocrats combo turns work.
//
// The sacrifice is a COST, so it happens on announce: the dies-
// triggers it causes (Blood Artist) go on the stack above this
// ability and resolve first.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edad60c6-80de-4033-af1b-a703ac332983",
		Name:         "Goblin Bombardment",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: deal 1 damage to any target",
			Cost:    effects.SacrificeACreature(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.DealDamage{
					Source: ctx.Source(),
					Target: item.Targets[0].ID,
					Amount: 1,
				}.Apply(ctx)
			},
		}},
	})
}
