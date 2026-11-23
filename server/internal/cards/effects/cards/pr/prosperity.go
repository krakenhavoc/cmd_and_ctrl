package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prosperity — Sorcery {X}{U} (EDHREC rank 3634):
//
//	"Each player draws X cards."
//
// The group hug X-draw. X is the announce-time value the cost engine
// charged; every seated player draws that many, APNAP from the
// active seat, each an ordinary draw so "whenever an opponent draws"
// payoffs see every card. X of zero draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c586312d-d04a-4bfb-bbb2-b41186ca178e",
		Name:         "Prosperity",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B05EachPlayerDraws(ctx.Game, item, ctx.X())
		},
	})
}
