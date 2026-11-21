package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blitzball Shot — Instant {1}{G} (Final Fantasy, #4462):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// A layer 7c pump and a layer 6 keyword grant on the one target, both
// until end of turn, and neither if the target is gone or no longer
// legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c99779b7-a271-416e-bdc2-0f9ba6ceb144",
		Name:         "Blitzball Shot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Blitzball Shot — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Blitzball Shot — trample"}.Apply(ctx)
		},
	})
}
