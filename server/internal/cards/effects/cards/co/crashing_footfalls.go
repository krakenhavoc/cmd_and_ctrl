package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crashing Footfalls — Sorcery with no mana cost (Special Guests, S82 BNS):
//
//	"Suspend 4—{G}
//	 Create two 4/4 green Rhino creature tokens with trample."
//
// Ancestral Vision's shape: no mana cost means it can never be cast from
// hand (CR 118.6), so suspend is the whole of its playability; the last time
// counter's free cast resolves it. Two Rhinos arrive together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b166d97-c2c1-4642-b151-9bb1ad32e362",
		Name:         "Crashing Footfalls",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(4, "{G}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("4/4 green Rhino with trample"), N: 2}.Apply(ctx)
		},
	})
}
