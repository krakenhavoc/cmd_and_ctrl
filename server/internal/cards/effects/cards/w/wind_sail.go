package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wind Sail — {1}{U} Sorcery (Starter 1999, S82 OLD batch 701, #5620):
//
//	"One or two target creatures gain flying until end of turn."
//
// Each target still legal as the spell resolves gains flying (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7fb6ff0-4555-4bd1-b00d-ba2ed1cb4afc",
		Name:         "Wind Sail",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures").WithCount(1, 2),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Wind Sail — flying"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
