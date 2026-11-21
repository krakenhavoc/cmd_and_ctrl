package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Realm Seekers — Creature — Elf Scout {4}{G}{G}, 0/0 (Foundations
// Jumpstart):
//
//	"This creature enters with X +1/+1 counters on it, where X is the
//	 total number of cards in all players' hands.
//	 {2}{G}, Remove a +1/+1 counter from this creature: Search your
//	 library for a land card, reveal it, put it into your hand, then
//	 shuffle."
//
// X is counted as the creature enters (Prime Speaker Zegana's shape). The
// card is on the stack, not in a hand, when that happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccb66cdc-909a-42c1-a10f-3886e9ed1936",
		Name:         "Realm Seekers",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return effects.CardsInAllHands(g)
			}, "Realm Seekers: enters with X +1/+1 counters, X the total number of cards in all players' hands"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Remove a +1/+1 counter from this creature: Search your library for a land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsLand() },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Realm Seekers — a land card, into your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
