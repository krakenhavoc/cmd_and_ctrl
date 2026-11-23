package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lull — Instant {1}{G} (Dominaria Remastered, #5948):
//
//	"Prevent all combat damage that would be dealt this turn.
//	 Cycling {2}"
//
// Fog with cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d53f4c24-8deb-485a-b03a-945211fdf10d",
		Name:         "Lull",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PreventAllCombatDamageThisTurn{Label: "Lull: prevent combat damage"}.Apply(ctx)
		},
	})
}
