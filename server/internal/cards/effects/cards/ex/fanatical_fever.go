package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fanatical Fever — Instant {2}{G}{G} (Ice Age, S82 OLD, #5636):
//
//	"Target creature gets +3/+0 and gains trample until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1231def-96fb-4408-a0c9-25d6932d7ead",
		Name:         "Fanatical Fever",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "trample")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Label: "Fanatical Fever — +3/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Fanatical Fever — trample"}.Apply(ctx)
		},
	})
}
