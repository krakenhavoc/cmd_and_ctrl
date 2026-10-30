package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nova Flame — Sorcery {X}{2}{R}{R} (Marvel Super Heroes Commander,
// #4455):
//
//	"Put X +1/+1 counters on target creature you control. It deals
//	 damage equal to its power to each other creature."
//
// Two instructions in printed order (Flame On!'s shape: the second waits
// for the counters' replacement prompt, CR 608.2c). The damage is dealt
// by the targeted creature, not by the spell, so its deathtouch or
// lifelink applies, and its power is read after the counters have landed
// and the layers have run. Every other creature on the battlefield is hit,
// the caster's own included, in one damage event. A target that left the
// battlefield (or stopped being a creature you control) in response makes
// the whole spell fizzle (CR 608.2b).
//
// Declared as a sweep of unknown size (Earthquake's X) for the bot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4310dfd-52eb-4c1f-b1ad-bfbc9b50fe7b",
		Name:         "Nova Flame",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok || ctx.X() <= 0 {
					return nil
				}
				return AddCounter{Target: id, Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, found := ctx.Game.LookupCardForEffect(id)
				if !found || !c.IsCreature() || c.CurrentPower() <= 0 {
					return nil
				}
				power := c.CurrentPower()
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, other := range MatchingBattlefield(ctx, And(Creature(), NotSelf(id))) {
						if err := (DealDamage{Source: id, Target: other.InstanceID, Amount: power}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
		),
	})
}
