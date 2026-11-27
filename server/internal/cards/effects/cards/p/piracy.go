package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Piracy — {U}{U} Sorcery (Starter 1999, S82 OLD batch 701, #5620):
//
//	"Until end of turn, you may tap lands you don't control for mana. Spend
//	 this mana only to cast spells."
//
// The engine seam of #6204 (game/tap_foreign_lands.go): a per-turn permission
// of the player, with the mana restricted to casting spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "131acf69-fa5c-4ce7-9e39-2677af4f8907",
		Name:         "Piracy",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB701Piracy(ctx.Game, item)
		},
	})
}
