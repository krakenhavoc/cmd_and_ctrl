package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mouth // Feed — Sorcery // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Mouth {2}{G}: "Create a 3/3 green Hippo creature token."
//	Feed {3}{G}: "Aftermath. Draw a card for each creature you control with
//	 power 3 or greater."
//
// Two registered faces. Feed counts as it resolves, through the layered view of
// power, creatures you control only; it is cast only from a graveyard and
// exiled as it resolves (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "980f6341-3c13-4405-8550-b838b228f788",
		Name:         "Mouth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("3/3 green Hippo"), N: 1}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "980f6341-3c13-4405-8550-b838b228f788#1",
		Name:         "Feed",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsCreature() && effects.B31CurrentPowerOf(ctx.Game, c.InstanceID) >= 3 {
					n++
				}
			}
			if n <= 0 {
				return nil
			}
			return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
		},
	})
}
