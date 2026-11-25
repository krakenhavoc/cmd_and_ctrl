package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ayula's Influence — Enchantment {G}{G}{G} (Modern Horizons):
//
//	"Discard a land card: Create a 2/2 green Bear creature token."
//
// Borborygmos Enraged's cost, a land card discarded from the hand (CR 701.8),
// at instant speed and with no limit: nothing resolves without the discard, so
// a hand with no land card can't activate it. The token is the catalog's 2/2
// green Bear.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25d955e3-2b98-4a75-bb51-6072000111be",
		Name:         "Ayula's Influence",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a land card: Create a 2/2 green Bear creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Bear"), N: 1}),
		}},
	})
}
