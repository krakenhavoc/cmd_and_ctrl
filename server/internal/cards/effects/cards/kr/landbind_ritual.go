package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Landbind Ritual — Sorcery {3}{W}{W} (Zendikar, S82 OLD b495):
//
//	"You gain 2 life for each Plains you control."
//
// The Plains are counted as the spell resolves (CR 608.2h), by land type, a word
// of the card's text (CR 612.2): a land that is a Plains through an effect counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63500b81-8fac-4208-91d3-0fe642899d87",
		Name:         "Landbind Ritual",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.CountYouControl(ctx, effects.HasSubtype(ctx.TextSubtype("Plains")))
			return effects.GainLife{Amount: 2 * n}.Apply(ctx)
		},
	})
}
