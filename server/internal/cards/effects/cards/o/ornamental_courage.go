package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ornamental Courage — Instant {G} (Kaladesh, #5539):
//
//	"Untap target creature. It gets +1/+3 until end of turn."
//
// Gift of Growth's two steps with a fixed pump (CR 608.2c): the untap, then
// one +1/+3 until end of turn, and a target that has gone skips both halves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1aaa85e1-4770-457f-a9af-3eeb56c71447",
		Name:         "Ornamental Courage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.UntapTarget{Target: id}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 3, Label: "Ornamental Courage — +1/+3"}.Apply(ctx)
			}),
		),
	})
}
