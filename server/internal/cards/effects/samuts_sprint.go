package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samut's Sprint — Instant {R} (War of the Spark, #5520):
//
//	"Target creature gets +2/+1 and gains haste until end of turn. Scry 1."
//
// The pump and the scry are steps, run in printed order (CR 608.2c).
// Samut, Tyrant Smasher's −1 is the same text and shares
// warPlusTwoOneAndHaste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d07821e-d34b-4441-a137-836ffce7127b",
		Name:         "Samut's Sprint",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Scry: 1, Answers: game.AnswerPump,
			Targets: game.ForTargets(BenefitUntilEOT(0, 2, 1, "haste"))},
		Targets: TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			WarPlusTwoOneAndHaste("Samut's Sprint — +2/+1 and haste"),
			Scry{N: 1},
		),
	})
}

// warPlusTwoOneAndHaste is "Target creature gets +2/+1 and gains haste
// until end of turn" as one step: the still-legal target gets both
// (CR 608.2b). Samut's Sprint and Samut, Tyrant Smasher's −1.
func WarPlusTwoOneAndHaste(label string) Instruction {
	return func(ctx *Context) error {
		return PumpAndGrantFirstTarget(2, 1, label, "haste")(ctx.Game, ctx.Item)
	}
}
