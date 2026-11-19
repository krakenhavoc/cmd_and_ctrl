package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "48b561dd-f23d-409a-87c8-e49902b2477b",
		Name:         "Traveler's Amulet",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Traveler's Amulet — a basic land card, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
