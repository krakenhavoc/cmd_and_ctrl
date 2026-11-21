package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Downsize — Instant {U} (Ravnica Remastered):
//
//	"Target creature you don't control gets -4/-0 until end of turn.
//	 Overload {2}{U} (You may cast this spell for its overload cost. If you
//	 do, change "target" in its text to "each.")"
//
// Overloaded it has no target and shrinks each creature you don't control
// as it resolves (CR 611.2c: the set is fixed then), as Corporeal
// Projection's overload does for its group.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46537bac-b1eb-4c02-ba43-6c6b16c7729c",
		Name:         "Downsize",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you don't control", OpponentControls()),
		Purpose:      ForTargets(HarmUntilEOT(0, -4, 0)),
		AlternativeCosts: []game.AlternativeCost{
			CostWithPurpose(Overload("{2}{U}"), game.Purpose{}),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			boost := BoostUntilEOT{Power: -4, Label: "Downsize — -4/-0"}
			if ctx.PaidAltCost("overload") {
				boost.Match = And(Creature(), OpponentControls())
				return boost.Apply(ctx)
			}
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			boost.Target = id
			return boost.Apply(ctx)
		},
	})
}
