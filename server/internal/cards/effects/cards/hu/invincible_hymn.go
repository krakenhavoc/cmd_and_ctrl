package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invincible Hymn — Sorcery {6}{W}{W} (Shards of Alara, S82 OLD batch 512,
// #5588):
//
//	"Count the number of cards in your library. Your life total becomes that
//	 number."
//
// The count is taken as the spell resolves; the life change is the
// difference (CR 119.5), through the ordinary life window.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc7d99e8-f9dc-47d3-b6c0-e9308d33c3c0",
		Name:         "Invincible Hymn",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p := ctx.PlayerByID(ctx.Controller())
			if p == nil || p.Library == nil {
				return nil
			}
			return effects.B31LifeBecomes(ctx, ctx.Controller(), p.Library.Size())
		},
	})
}
