package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f6a07af3-e54b-4dcb-ae2a-3dfdac386bab",
		Name:            "Wild-Field Scarecrow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Search your library for up to two basic land cards, reveal them, put them into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     2,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Wild-Field Scarecrow — up to two basic land cards, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
