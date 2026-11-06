package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maximum Overdrive — Instant {1}{B} (Aetherdrift, #4464):
//
//	"Put a +1/+1 counter on target creature. It gains deathtouch and
//	 indestructible until end of turn."
//
// Fleeting Flight's two steps with the keywords changed: the counter
// first, then the keywords, the second step waiting for any prompt the
// counter asks (a replacement order, CR 608.2c). The counter stays; the
// keywords end with the turn. A target gone by resolution gets neither
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0e85b74-12fb-4d63-853e-7a3f41f731b4",
		Name:         "Maximum Overdrive",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerSurvive | game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch", "indestructible"},
						Label: "Maximum Overdrive — deathtouch and indestructible"}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
