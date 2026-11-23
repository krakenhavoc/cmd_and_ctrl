package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Join the Dance — Sorcery {G}{W} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Create two 1/1 white Human creature tokens.
//	 Flashback {3}{G}{W}"
//
// Army of the Damned's shape: the flashback cast from the graveyard
// keeps sorcery timing and exiles the card however it leaves the stack
// (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d8fc015c-7e6d-441d-bbeb-98ca74b21314",
		Name:             "Join the Dance",
		Purpose:          game.Purpose{Tokens: 2},
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{G}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Human"), N: 2}.Apply(ctx)
		},
	})
}
