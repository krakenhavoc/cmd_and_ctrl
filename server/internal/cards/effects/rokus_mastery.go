package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roku's Mastery — Instant {X}{R}{R} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Roku's Mastery deals X damage to target creature. If X is 4 or more,
//	 scry 2."
//
// X is the announced value (CR 107.3). The scry reads X, not the damage
// dealt, so a prevented or partly absorbed blow still scries at 4 or
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5659c45-e064-4690-b90a-24ffe3793d4d",
		Name:         "Roku's Mastery",
		XMatters:     true,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: ctx.X()}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.X() < 4 {
				return nil
			}
			return Scry{N: 2}.Apply(ctx)
		},
	})
}
