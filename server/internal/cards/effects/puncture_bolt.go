package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Puncture Bolt — Instant {1}{R} (Shadowmoor, #5591):
//
//	"Puncture Bolt deals 1 damage to target creature. Put a -1/-1
//	 counter on that creature."
//
// Two instructions in order (CR 608.2c): 1 damage, then a -1/-1 counter on
// the same creature. State-based actions wait for the whole spell, so a
// 2/2 ends 1/1 with 1 damage marked and dies (CR 704.5g).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8e53740-c633-42f8-9e02-3434d6dc0b30",
		Name:         "Puncture Bolt",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 1)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
		},
	})
}
