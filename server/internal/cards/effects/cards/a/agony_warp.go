package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agony Warp — Instant {U}{B} (S82 GAP, #5948):
//
//	"Target creature gets -3/-0 until end of turn.
//	 Target creature gets -0/-3 until end of turn."
//
// Two target clauses, each re-checked on its own at resolution (CR 608.2b):
// the same creature may be both, and one illegal target does not stop the other.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "377c217b-ddad-4e1a-9122-f347113d3dfa",
		Name:         "Agony Warp",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, 0), effects.HarmUntilEOT(1, 0, -3)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature (gets -3/-0)"),
			effects.TargetCreature("target creature (gets -0/-3)"),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: -3, Label: "Agony Warp — -3/-0"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Toughness: -3, Label: "Agony Warp — -0/-3"}.Apply(ctx)
			}),
		),
	})
}
