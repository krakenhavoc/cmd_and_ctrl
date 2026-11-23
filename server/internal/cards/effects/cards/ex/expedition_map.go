package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expedition Map — Artifact {1} (EDHREC rank 457):
//
//	"{2}, {T}, Sacrifice this artifact: Search your library for a
//	 land card, reveal it, put it into your hand, then shuffle."
//
// The one-mana land tutor — Cabal Coffers, Urborg, Field of the
// Dead, Gaea's Cradle, whichever land the deck is built around.
// Three cost components (Myriad Landscape's), a search to HAND with
// the reveal the printed text asks for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fcf50cd-e6d0-4516-850f-d42ee75dcc3a",
		Name:         "Expedition Map",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search your library for a land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.B03IsLandCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Expedition Map — a land card, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
