package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unending Whisper — Sorcery {U} (#3931):
//
//	"Draw a card.
//	 Harmonize {5}{U}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c5da5c31-0cf1-4393-8701-eda7880ecebc",
		Name:             "Unending Whisper",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Harmonize("{5}{U}")},
		Purpose:          game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
