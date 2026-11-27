package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Research — Enchantment {2}{R} (Apocalypse, S82 OLD, #5614):
//
//	"{1}{W}: Search your library for an enchantment card and reveal that card.
//	 Put it into your hand, then discard a card at random. Then shuffle.
//	 {1}{U}: Search your library for an instant card and reveal that card. Put
//	 it into your hand, then discard a card at random. Then shuffle."
//
// Gamble's chained random discard behind a revealing search of one card type.
// The search returns while its prompt is open, so the discard is the search's
// tail and takes a card from the hand the found card is in; the shuffle goes
// with the search, and nothing can see a library shuffle before a discard.
//
// No simplification.
func init() {
	tutor := func(what string, match func(game.Card) bool) func(g *game.Game, item *game.StackItem) error {
		return func(g *game.Game, item *game.StackItem) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: match,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Wild Research — search your library for " + what,
				ThenTail:  game.KeyedTail{Ref: effects.ThenDiscardAtRandom, Player: item.Controller, N: 1},
			}.Apply(effects.NewContext(g, item))
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "d4721c7e-1a55-42cf-99c9-9118e386f48b",
		Name:         "Wild Research",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}: Search your library for an enchantment card and reveal that card. Put it into your hand, then discard a card at random. Then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1, Discards: 1},
				Cost:    effects.ManaCost("{1}{W}"),
				Effect:  tutor("an enchantment card", func(c game.Card) bool { return c.IsEnchantment() }),
			},
			{
				Label:   "{1}{U}: Search your library for an instant card and reveal that card. Put it into your hand, then discard a card at random. Then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1, Discards: 1},
				Cost:    effects.ManaCost("{1}{U}"),
				Effect:  tutor("an instant card", func(c game.Card) bool { return c.IsInstant() }),
			},
		},
	})
}
