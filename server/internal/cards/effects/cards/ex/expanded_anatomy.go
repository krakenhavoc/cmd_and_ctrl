package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expanded Anatomy — Sorcery — Lesson {3} (Strixhaven: School of Mages,
// #5502):
//
//	"Put two +1/+1 counters on target creature. It gains vigilance until
//	 end of turn."
//
// Angelfire Ignition's two counters and keyword grant, with vigilance
// alone. Being a Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bc5dfbb-98ec-4ff1-8ba2-e2594c53e6d9",
		Name:         "Expanded Anatomy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Expanded Anatomy — vigilance"}.Apply(ctx)
		},
	})
}
