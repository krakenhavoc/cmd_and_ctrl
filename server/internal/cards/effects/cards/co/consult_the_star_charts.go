package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consult the Star Charts — Instant {1}{U} (Edge of Eternities):
//
//	"Kicker {1}{U}
//	 Look at the top X cards of your library, where X is the number of
//	 lands you control. Put one of those cards into your hand. If this
//	 spell was kicked, put two of those cards into your hand instead. Put
//	 the rest on the bottom of your library in a random order."
//
// X is counted once, as the spell resolves (the rulings). The pick is
// mandatory (the floor is as many cards as the clause names, or the whole
// look if it is smaller) and unrevealed; the rest go to the bottom in a
// random order (takeRestBottomRandom). A copy of a kicked spell is kicked
// too, which ctx.KickedTimes reads from the item.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e921839f-9d91-41a9-bc89-016af3c757aa",
		Name:          "Consult the Star Charts",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{U}")},
		Purpose:       game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			lands := effects.CountControlled(ctx.Game, player, func(c game.Card) bool { return c.IsLand() })
			take := 1
			if ctx.KickedTimes() > 0 {
				take = 2
			}
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, lands),
				Max:    take,
				Label:  "Consult the Star Charts — put the chosen cards into your hand",
				Key:    effects.TakeRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
