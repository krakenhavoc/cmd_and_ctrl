package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Journeyer's Kite — Artifact {2} (Champions of Kamigawa, #5603):
//
//	"{3}, {T}: Search your library for a basic land card, reveal it, put
//	 it into your hand, then shuffle."
//
// A repeatable land tutor to hand; the card found is revealed (CR
// 701.20a). It is not a creature, so it can be tapped the turn it
// arrives (CR 302.6 applies to creatures only).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10aab5bc-5758-443b-ba4c-49f6c6d91262",
		Name:         "Journeyer's Kite",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Journeyer's Kite — choose a basic land card to put into your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
