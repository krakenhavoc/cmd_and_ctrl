package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assault Strobe — Sorcery {R} (Scars of Mirrodin, #5577):
//
//	"Target creature gains double strike until end of turn."
//
// Critical Hit's grant at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9bf4ee0f-6e28-4793-b251-f81ee444a119",
		Name:         "Assault Strobe",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Assault Strobe — double strike"}.Apply(ctx)
		},
	})
}
