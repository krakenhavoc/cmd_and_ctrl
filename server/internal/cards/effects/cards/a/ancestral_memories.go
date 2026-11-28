package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancestral Memories — Sorcery {2}{U}{U}{U} (Portal, S82 OLD batch 781, #5630):
//
//	"Look at the top seven cards of your library. Put two of them into your
//	 hand and the rest into your graveyard."
//
// Bitter Revelation's look with seven cards and no life loss: a private look,
// a mandatory pick of two for the hand, the rest put into the graveyard (not
// milled). With fewer cards in the library you take what is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95a2802a-2621-40c3-84f8-51e8aad7b6f0",
		Name:         "Ancestral Memories",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
				Max:    2,
				Label:  "Ancestral Memories — put two of them into your hand and the rest into your graveyard",
				Key:    ancestralMemoriesTake,
			}.Apply(ctx)
		},
	})
}

// ancestralMemoriesTake is the look's continuation: the rest into the
// graveyard. An on-disk key: never renamed, never reused.
var ancestralMemoriesTake = effects.RegisterTake("choose-cards/s82-old-b781-ancestral-memories-take",
	effects.TakeDef{Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		return effects.TakeRestIntoGraveyard(g, res)
	}})
