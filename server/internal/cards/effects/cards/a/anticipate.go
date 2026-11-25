package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anticipate — Instant {1}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Look at the top three cards of your library. Put one of them into your
//	 hand and the rest on the bottom of your library in any order."
//
// Impulse's dig with three cards: a private look, one mandatory pick, the
// rest bottomed in an order the player chooses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a97da5a9-17fc-4367-81e1-17ff81601a63",
		Name:         "Anticipate",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LookAtTopThenTakeOneRestOnBottom(3, "Anticipate — put one of them into your hand")(ctx.Game, item)
		},
	})
}
