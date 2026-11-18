package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seismic Wave — Instant {2}{R} (Kamigawa: Neon Dynasty, S82 #5493):
//
//	"Seismic Wave deals 2 damage to any target and 1 damage to each
//	 nonartifact creature target opponent controls."
//
// Two targets, so two clauses (CR 601.2c): any target, and an opponent.
// Each is checked on its own at resolution (CR 608.2b), so if one is
// illegal the other still happens. The 1 damage hits the opponent's
// nonartifact creatures, the set fixed before any is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a56e2ef-75dd-47d5-b61c-11faaba8cd9a",
		Name:         "Seismic Wave",
		Completeness: CompletenessFull,
		Targets:      Clauses(TargetAny(), TargetPlayer("target opponent", Opponent())),
		Purpose: game.Purpose{
			Targets: game.ForTargets(DamageToTarget(0, 2)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetPlayer {
				return neoB15DamageEachNonartifactCreatureControlledBy(ctx, t.ID, 1)
			}
			return nil
		},
	})
}
