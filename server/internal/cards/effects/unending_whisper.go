package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unending Whisper — Sorcery {U} (#3931):
//
//	"Draw a card.
//	 Harmonize {5}{U}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c5da5c31-0cf1-4393-8701-eda7880ecebc",
		Name:             "Unending Whisper",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{5}{U}")},
		Purpose:          game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
