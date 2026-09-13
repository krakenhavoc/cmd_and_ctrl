package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "edad60c6-80de-4033-af1b-a703ac332983",
		Name:         "Goblin Bombardment",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: deal 1 damage to any target",
			Cost:    SacrificeACreature(),
			Targets: TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := NewContext(g, item)
				return DealDamage{
					Source: ctx.Source(),
					Target: item.Targets[0].ID,
					Amount: 1,
				}.Apply(ctx)
			},
		}},
	})
}
