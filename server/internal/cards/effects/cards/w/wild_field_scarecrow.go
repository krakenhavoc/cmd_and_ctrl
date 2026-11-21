package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild-Field Scarecrow — Artifact Creature — Scarecrow {3}, 1/4 (Innistrad
// Remastered, #5948):
//
//	"Defender
//	 {2}, Sacrifice this creature: Search your library for up to two basic land
//	 cards, reveal them, put them into your hand, then shuffle."
//
// The sacrifice is a cost, so the search happens whatever becomes of the
// Scarecrow. A search with a stated quality may find fewer than two, or none
// (CR 701.23b); each basic land card found is revealed into the hand before the
// shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6a07af3-e54b-4dcb-ae2a-3dfdac386bab",
		Name:            "Wild-Field Scarecrow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Search your library for up to two basic land cards, reveal them, put them into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     2,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Wild-Field Scarecrow — up to two basic land cards, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
