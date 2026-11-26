package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kin-Tree Invocation — Sorcery {B}{G} (Khans of Tarkir):
//
//	"Create an X/X black and green Spirit Warrior creature token, where X is
//	 the greatest toughness among creatures you control."
//
// X is read as it resolves (CR 608.2c) from the toughness each creature has
// then (Glint Weaver's greatest toughness), and fixes the token's size as it is
// created (Grakmaw's X/X Hydra); with no creatures it is 0/0 and dies as a
// state-based action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbb74ff8-7381-4de9-bf62-e5d181a187bc",
		Name:         "Kin-Tree Invocation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := effects.GlintWeaverGreatestToughnessControlledBy(ctx.Game, ctx.Controller())
			tok := ctx.TextToken(effects.TokenCard("0/0 black and green Spirit Warrior"))
			tok.Power, tok.Toughness = x, x
			return ctx.Game.CreateTokenForEffect(item.Controller, tok, 1)
		},
	})
}
