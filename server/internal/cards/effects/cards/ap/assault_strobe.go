package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assault Strobe — Sorcery {R} (Scars of Mirrodin, #5577):
//
//	"Target creature gains double strike until end of turn."
//
// Critical Hit's grant at sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bf4ee0f-6e28-4793-b251-f81ee444a119",
		Name:         "Assault Strobe",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Assault Strobe — double strike"}.Apply(ctx)
		},
	})
}
