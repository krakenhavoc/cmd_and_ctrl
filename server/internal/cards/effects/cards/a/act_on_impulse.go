package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Act on Impulse — Sorcery {2}{R} (Jumpstart 2022, #5948):
//
//	"Exile the top three cards of your library. Until end of turn, you may
//	 play those cards."
//
// Light Up the Stage's exile with three cards and the turn's own window:
// the grant is "play", so a land is playable (it still needs a land drop).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19ad338a-c89f-406d-ad1a-222c22044e1b",
		Name:         "Act on Impulse",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 3}.Apply(ctx)
		},
	})
}
