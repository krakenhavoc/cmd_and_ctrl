package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colossal Might — Instant {R}{G} (Alara Reborn, #5586):
//
//	"Target creature gets +4/+2 and gains trample until end of turn."
//
// Blitzball Shot's shape at +4/+2: a layer 7c boost and a layer 6 trample
// grant, both ending in the cleanup step. A target that left in response
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d8e39c9-e04d-4328-af66-db6e1241ac06",
		Name:         "Colossal Might",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 2, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 4, Toughness: 2, Label: "Colossal Might — +4/+2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Colossal Might — trample"}.Apply(ctx)
		},
	})
}
