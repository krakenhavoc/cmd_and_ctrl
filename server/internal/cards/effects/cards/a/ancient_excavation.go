package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Excavation — Instant {2}{U}{B} (Commander 2016, #5537):
//
//	"Draw cards equal to the number of cards in your hand, then discard a card
//	 for each card drawn this way.
//	 Basic landcycling {2} ({2}, Discard this card: Search your library for a
//	 basic land card, reveal it, put it into your hand, then shuffle.)"
//
// The hand is counted as the spell resolves (it is already on the stack, so it
// is not in the count); the discard opens only once every draw has resolved
// (LootOne, CR 608.2c), so a drawn card is a legal discard. Basic landcycling
// is the shared hand ability (Fiery Fall).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46db3dc2-9eb1-4be3-96f9-0013119fcd97",
		Name:         "Ancient Excavation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.CardsInHand(ctx.Game, item.Controller)
			if n <= 0 {
				return nil
			}
			return effects.LootOne(ctx.Game, item, n)
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
