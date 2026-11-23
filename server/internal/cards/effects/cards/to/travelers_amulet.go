package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traveler's Amulet — Artifact {1} (Innistrad Remastered, #5948):
//
//	"{1}, Sacrifice this artifact: Search your library for a basic land card,
//	 reveal it, put it into your hand, then shuffle."
//
// World Map's first ability without the {T}: the sacrifice is a cost, so the
// search happens whatever becomes of the Amulet, and the basic land card found
// (any basic, by supertype) is revealed into the hand before the shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48b561dd-f23d-409a-87c8-e49902b2477b",
		Name:         "Traveler's Amulet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Traveler's Amulet — a basic land card, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
