package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glintwing Invoker — Creature — Human Wizard Mutant {4}{U}, 3/3
// (Dominaria Remastered, #5948):
//
//	"{7}{U}: This creature gets +3/+3 and gains flying until end of turn."
//
// The pump and the keyword are two turn-scoped effects on the source, each
// a different kind of thing (layer 7c and layer 6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8a9bdaa-55ef-4baa-9dfa-9068583360f6",
		Name:         "Glintwing Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{U}: This creature gets +3/+3 and gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3, Keywords: []string{"flying"}}},
			Cost:    effects.ManaCost("{7}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 3, Toughness: 3, Label: "Glintwing Invoker — +3/+3"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"}, Label: "Glintwing Invoker — flying"}.Apply(ctx)
			},
		}},
	})
}
