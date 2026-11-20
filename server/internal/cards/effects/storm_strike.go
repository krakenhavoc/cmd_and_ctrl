package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f0748ff2-94b1-45a9-8d56-8459bd86704a",
		Name:         "Storm Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      WithRiders(ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")), game.Purpose{Scry: 1}),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return pumpAndGrantFirstTarget(1, 0, "Storm Strike — +1/+0 and first strike", "first strike")(ctx.Game, ctx.Item)
			}),
			Scry{N: 1},
		),
	})
}
