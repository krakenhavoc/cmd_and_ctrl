package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Infusion — Sorcery {1}{B} (Strixhaven: School of Mages, #5502):
//
//	"Put two +1/+1 counters on target creature. It gains lifelink until
//	 end of turn."
//
// Angelfire Ignition's two counters and keyword grant, with lifelink
// alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60f412bb-27e9-4883-8862-96eae96dfee3",
		Name:         "Essence Infusion",
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
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Essence Infusion — lifelink"}.Apply(ctx)
		},
	})
}
