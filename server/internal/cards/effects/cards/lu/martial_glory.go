package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Martial Glory — Instant {R}{W} (S82 GAP, #5948):
//
//	"Target creature gets +3/+0 until end of turn.
//	 Target creature gets +0/+3 until end of turn."
//
// Two target clauses, each re-checked on its own (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8aa12f8-8922-41ed-8cdb-219c9e4b9e36",
		Name:         "Martial Glory",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0), effects.BenefitUntilEOT(1, 0, 3)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature (gets +3/+0)"),
			effects.TargetCreature("target creature (gets +0/+3)"),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 3, Label: "Martial Glory — +3/+0"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Toughness: 3, Label: "Martial Glory — +0/+3"}.Apply(ctx)
			}),
		),
	})
}
