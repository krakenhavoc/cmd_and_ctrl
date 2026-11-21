package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flow of Ideas — Sorcery {5}{U} (Ravnica: City of Guilds, #5600):
//
//	"Draw a card for each Island you control."
//
// Flow of Knowledge's count without the discard: every land its caster
// controls with the Island subtype, read as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd397a32-06b0-4530-b075-412e2881aa53",
		Name:         "Flow of Ideas",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B33IslandsControlled(ctx.Game, item.Controller, ctx.TextSubtype("Island"))
			return effects.DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
