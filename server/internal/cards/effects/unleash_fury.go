package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unleash Fury — Instant {1}{R}:
//
//	"Double the power of target creature until end of turn."
//
// Bulk Up without the flashback. CR 701.10b: the creature gets +X/+0
// until end of turn, X its power as the spell resolves, counters and
// other effects included; a negative power doubles downward (CR
// 701.10c). DoublePowerUntilEOT (doubling.go) is the whole body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4eaa100c-f1ef-4a2b-9370-5dad7f3a95f2",
		Name:         "Unleash Fury",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return DoublePowerUntilEOT(ctx, id, "Unleash Fury — double power")
		},
	})
}
