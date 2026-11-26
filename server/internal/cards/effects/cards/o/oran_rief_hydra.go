package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oran-Rief Hydra — Creature — Hydra {4}{G}{G}, 5/5 (Battle for Zendikar):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 this creature. If that land is a Forest, put two +1/+1 counters on this
//	 creature instead."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b065a17-14a4-433f-a6c8-2c8212f505c7",
		Name:            "Oran-Rief Hydra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Oran-Rief Hydra — +1/+1 counter, two if that land is a Forest",
				func(g *game.Game, item *game.StackItem) error {
					n := 1
					if effects.S82OldB309EnteredLandIs(g, item, "Forest") {
						n = 2
					}
					ctx := effects.NewContext(g, item)
					return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
	})
}
