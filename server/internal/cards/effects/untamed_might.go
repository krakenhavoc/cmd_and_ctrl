package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Untamed Might — Instant {X}{G} (Scars of Mirrodin, #5577):
//
//	"Target creature gets +X/+X until end of turn."
//
// Enrage's shape on both sides: X is the announced X (CR 107.3a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b6dd710-f7f2-4fc8-9b66-b48c7737b135",
		Name:         "Untamed Might",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(PumpXUntilEOT(0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: ctx.X(), Toughness: ctx.X(), Label: "Untamed Might — +X/+X until end of turn"}.Apply(ctx)
		},
	})
}
