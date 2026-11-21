package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Offering to Asha — Instant {2}{W}{U} (Alara Reborn, #5586):
//
//	"Counter target spell unless its controller pays {4}. You gain 4
//	 life."
//
// Convolute's tax, then Sacred Nectar's life gain as a second step, so the
// gain waits for the payer's answer (CR 608.2c). The life is gained
// whether or not the tax is paid; a target that left in response is
// skipped and the life is still gained (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d061972e-6f4f-41a7-966a-88839f105f19",
		Name:         "Offering to Asha",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{4}",
					Question: "Offering to Asha — pay {4} or your spell is countered",
				}.Apply(ctx)
			}),
			effects.GainLife{Amount: 4},
		),
	})
}
