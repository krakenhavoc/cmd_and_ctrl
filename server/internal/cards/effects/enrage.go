package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enrage — Instant {X}{R} (Scourge):
//
//	"Target creature gets +X/+0 until end of turn."
//
// Hatred's +X/+0 with X paid in mana instead of life: X is announced with
// the cast (CR 601.2b) and read as the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f6e66d5-4f27-485b-999d-ee55c1e218b9",
		Name:         "Enrage",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(PowerXUntilEOT(0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: ctx.X(), Label: "Enrage — +X/+0 until end of turn"}.Apply(ctx)
		},
	})
}
