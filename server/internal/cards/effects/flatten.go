package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flatten — Instant {3}{B} (Dragons of Tarkir,
// #5551):
//
//	"Target creature gets -4/-4 until end of turn."
//
// Dark Deed's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc5d5084-77de-4dc5-8769-300eafcbfee7",
		Name:         "Flatten",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, -4)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: -4, Toughness: -4, Label: "Flatten — -4/-4 until end of turn"}.Apply(ctx)
		},
	})
}
