package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infuriate — Instant {R}:
//
//	"Target creature gets +3/+2 until end of turn."
//
// Giant Growth's shape with printed +3/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d23119ca-73d1-4a48-a8d0-dab8919d891f",
		Name:         "Infuriate",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: 3, Toughness: 2, Label: "Infuriate — +3/+2"}.Apply(ctx)
		},
	})
}
