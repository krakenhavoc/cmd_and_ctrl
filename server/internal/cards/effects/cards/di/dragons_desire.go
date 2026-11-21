package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon's Desire — Sorcery {2}{R}{R} (The Hobbit):
//
//	"Add {R} for each artifact your opponents control."
//
// Mana Geyser's shape over a different count. The artifacts are counted on
// resolution, across every opponent's side of the battlefield (tokens
// included); none adds nothing, since AddMana ignores an empty count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caef715c-3ca6-4f56-9dd5-aebce9cd0621",
		Name:         "Dragon's Desire",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.AddMana{
				Player: item.Controller,
				Amount: effects.ManaN(effects.HobB7ArtifactsByOpponents(ctx.Game, item.Controller), "R"),
			}.Apply(ctx)
		},
	})
}
