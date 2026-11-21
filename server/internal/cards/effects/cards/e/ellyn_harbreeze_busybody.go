package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ellyn Harbreeze, Busybody — Legendary Creature — Human Peasant {3}{W},
// 2/4 (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"{T}: Look at the top X cards of your library, where X is the number of
//	 tokens you created this turn. Put one of those cards into your hand and
//	 the rest on the bottom of your library in a random order.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// X is the controller's token-creation tally for the turn
// (PlayerTurnTally.TokensCreated: tokens created, not tokens on the
// battlefield), read as the ability resolves. The look is Astor's
// LookAtTopOfLibrary; the take is mandatory when at least one card was looked
// at, and the rest go to the bottom in a random order (TakeRestBottomRandom).
// Choose a Background is a deck-construction rule (CR 702.124k) that
// internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca5ecc26-5ed8-4466-bb8f-df2b39bddd18",
		Name:         "Ellyn Harbreeze, Busybody",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Look at the top X cards of your library, where X is the number of tokens you created this turn. Put one of those cards into your hand and the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := g.TurnTallyFor(item.Controller).TokensCreated
				if x <= 0 {
					return nil
				}
				return effects.TakeFromLibraryToHand{
					Player: item.Controller,
					Cards:  g.LookAtTopOfLibraryForEffect(item.Controller, x),
					Max:    1,
					Label:  "Ellyn Harbreeze, Busybody — put one of those cards into your hand",
					Key:    effects.TakeRestBottomRandom,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
