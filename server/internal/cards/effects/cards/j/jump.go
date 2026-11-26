package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jump — Instant {U} (Magic 2010, #5585):
//
//	"Target creature gains flying until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7518456-45ed-41d4-bd3c-5aacea28eb35",
		Name:         "Jump",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Jump — flying"}.Apply(ctx)
		},
	})
}
