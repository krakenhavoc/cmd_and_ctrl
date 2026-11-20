package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stirring Address — Instant {1}{W} (Modern Horizons, #5519):
//
//	"Target creature you control gets +2/+2 until end of turn.
//	 Overload {5}{W} (You may cast this spell for its overload cost. If you
//	 do, change "target" in its text to "each.")"
//
// Fangs of Kalonia's overload shape on a pump: the overloaded cast has no
// target and gives +2/+2 to each creature you control at resolution
// (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45a60f56-35e4-46e5-838f-8b4407a67cd6",
		Name:         "Stirring Address",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{5}{W}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Toughness: 2, Label: "Stirring Address — +2/+2"}.Apply(ctx)
			}
			return boostTheTargetUntilEOT(2, 2, "Stirring Address — +2/+2")(item, ctx)
		},
	})
}
