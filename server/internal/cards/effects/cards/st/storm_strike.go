package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Strike — Instant {R} (Ravnica Allegiance, #5521):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn.
//	 Scry 1."
//
// Kindled Heroism's text: the pump and the first strike on the target,
// then scry 1 as a second step, so the scry still happens after a prompt
// the first step asked. An illegal target counters the whole spell, scry
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0748ff2-94b1-45a9-8d56-8459bd86704a",
		Name:         "Storm Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.WithRiders(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")), game.Purpose{Scry: 1}),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.PumpAndGrantFirstTarget(1, 0, "Storm Strike — +1/+0 and first strike", "first strike")(ctx.Game, ctx.Item)
			}),
			effects.Scry{N: 1},
		),
	})
}
