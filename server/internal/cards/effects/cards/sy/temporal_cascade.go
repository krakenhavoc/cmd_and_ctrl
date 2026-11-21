package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temporal Cascade — Sorcery {5}{U}{U}:
//
//	"Choose one —
//	 • Each player shuffles their hand and graveyard into their library.
//	 • Each player draws seven cards.
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). The two bullets are Echo of Eons' two
// phases taken apart: the shuffle is echoOfEonsShuffleEach with nobody
// left to draw, one simultaneous tuck per player (a commander in a hand
// or graveyard gets its CR 903.9b offer), and the draw is
// echoOfEonsDrawEach. Entwined, every player shuffles before anyone
// draws (CR 702.42b); a commander prompt in the shuffle holds the draw
// until it is answered (#2789).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c65dc9cb-3aad-4d15-a8c0-c0c142058d45",
		Name:          "Temporal Cascade",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Each player shuffles their hand and graveyard into their library.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.EchoOfEonsShuffleEach(ctx.Game, effects.TablePlayers(ctx), nil)
				}),
			effects.ModeDoing("Each player draws seven cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.EchoOfEonsDrawEach(ctx.Game, effects.TablePlayers(ctx))
				}),
		)),
	})
}
