package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Braidwood Sextant — Artifact {1} (Urza's Destiny, #5621):
//
//	"{2}, {T}, Sacrifice this artifact: Search your library for a basic
//	 land card, reveal that card, put it into your hand, then shuffle."
//
// Expedition Map narrowed to a basic land card. The costs are paid at
// activation, so the Sextant is in the graveyard while the search
// resolves; the found card is revealed to the table before it goes to
// hand, and the library is shuffled even when nothing is found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b44816b4-44ef-446d-a254-b4b7bc015a42",
		Name:         "Braidwood Sextant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search your library for a basic land card, reveal that card, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Braidwood Sextant — a basic land card, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
