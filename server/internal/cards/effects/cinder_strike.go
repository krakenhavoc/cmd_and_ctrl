package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cinder Strike — {R} Sorcery:
//
//	"As an additional cost to cast this spell, you may blight 1. (You
//	 may put a -1/-1 counter on a creature you control.)
//	 Cinder Strike deals 2 damage to target creature. It deals 4 damage
//	 to that creature instead if this spell's additional cost was paid."
//
// #1703's plainest blight card: the optional cost is OptionalBlight(1),
// paid at CR 601.2h onto a creature the caster names, and the
// resolution reads the paid record. No simplification.
func init() {
	Register(Spec{
		OracleID:      "54421e69-d79e-4c2e-8ce6-96994d168835",
		Name:          "Cinder Strike",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalBlight(1)},
		Targets:       TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			amount := 2
			if ctx.BlightPaid() {
				amount = 4
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
