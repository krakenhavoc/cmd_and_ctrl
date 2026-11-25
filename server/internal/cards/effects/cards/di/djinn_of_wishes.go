package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Djinn of Wishes — Creature — Djinn {3}{U}{U}, 4/4:
//
//	"Flying
//	 This creature enters with three wish counters on it.
//	 {2}{U}{U}, Remove a wish counter from this creature: Reveal the top card
//	 of your library. You may play that card without paying its mana cost. If
//	 you don't, exile it."
//
// The counters are Ring of Three Wishes' entry replacement and the removal is
// part of the cost (CR 602.2b). The reveal is public. Playing is the free-play
// offer of Ziatora's Envoy over the revealed top card (game.FreePlayFromTop,
// #3801): a land the player can play now is a land play, any spell gets a
// free cast from the library during the resolution (CR 608.2g), whatever its
// mana value. It differs from the Envoy only in the decline: DeclineToExile
// sends the card that is not played, or cannot be (a land with no land drop,
// or on another player's turn), to exile instead of the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a34c3012-f787-4998-93c7-89c4decaa1ba",
		Name:            "Djinn of Wishes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("wish", 3, "Djinn of Wishes: enters with three wish counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{U}, Remove a wish counter from this creature: Reveal the top card of your library. You may play that card without paying its mana cost. If you don't, exile it.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}{U}"), effects.RemoveCountersFromThis("wish", 1)),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
					Player:         item.Controller,
					Source:         item.SourceCardID,
					SourceName:     "Djinn of Wishes",
					MaxManaValue:   1 << 20,
					Land:           true,
					Reveal:         true,
					DeclineToExile: true,
				})
			},
		}},
	})
}
