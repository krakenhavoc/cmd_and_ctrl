package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burst of Speed — Sorcery {R} (Magic 2010, #5585):
//
//	"Creatures you control gain haste until end of turn."
//
// The creatures you control as it resolves gain haste (CR 611.2c), so one
// that enters later this turn does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91e33674-b13a-41de-8998-4b57016ae42f",
		Name:         "Burst of Speed",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"haste"},
				Label:    "Burst of Speed — creatures you control gain haste",
			}.Apply(ctx)
		},
	})
}
