package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psychotic Fury — Instant {1}{R} (Dissension, #5598):
//
//	"Target multicolored creature gains double strike until end of turn.
//	 Draw a card."
//
// Assault Strobe's grant on a multicolored target, then a draw. A target
// that is no longer multicolored as the spell resolves counters it (CR
// 608.2b), and the draw goes with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bf982e9-5868-46bf-81e3-f29ada33d169",
		Name:         "Psychotic Fury",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      TargetCreature("target multicolored creature", Multicolored()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Psychotic Fury — double strike"}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
