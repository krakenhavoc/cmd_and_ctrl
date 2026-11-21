package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howl from Beyond — Instant {X}{B} (Dominaria Remastered, #5948):
//
//	"Target creature gets +X/+0 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "403cf6ae-48a9-4ea9-894e-7135cfca4e1b",
		Name:         "Howl from Beyond",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(PowerXUntilEOT(0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: ctx.X(), Label: "Howl from Beyond — +X/+0"}.Apply(ctx)
		},
	})
}
