package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Immolation — Sorcery {3}{R}{R}:
//
//	"As an additional cost to cast this spell, blight X. X can't be
//	 greater than the greatest toughness among creatures you control.
//	 (Put X -1/-1 counters on a creature you control.)
//	 Soul Immolation deals X damage to each opponent and each creature
//	 they control."
//
// #2174: BlightXCost(). X is announced with the cast, capped by the
// greatest toughness among the caster's creatures, and the X -1/-1
// counters land on the one creature the caster names at CR 601.2h —
// with the spell already on the stack, so a creature that dies of them
// is gone before this resolves and the cost stays paid. The damage
// reads the same announced X (CR 107.3i). Damage is the snapshot-then-
// strike sweep Tectonic Hazard uses, so nothing is hit twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "338747e8-bed8-4e60-8b19-5c2b80799477",
		Name:           "Soul Immolation",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: BlightXCost(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, opp := range ctx.Opponents() {
					if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: x}).Apply(ctx); err != nil {
						return err
					}
				}
				return damageEachMatching(ctx, And(Creature(), OpponentControls()), x)
			})
		},
	})
}
