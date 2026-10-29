package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hobbit Hole — Land (The Hobbit, #4454):
//
//	"{T}, Sacrifice this land: Search your library for a basic land card,
//	 put it onto the battlefield tapped, then shuffle.
//	 Halflingcycling {4} ({4}, Discard this card: Search your library for
//	 a Halfling card, reveal it, put it into your hand, then shuffle.)"
//
// Evolving Wilds' fetch (fetchBasicTapped) with a typecycling ability
// (CR 702.29e) that searches for a Halfling card of any type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17492186-9814-4c41-8111-1f000a96c212",
		Name:         "Hobbit Hole",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
				Cost:    Plus(TapCost(), SacrificeThis()),
				Effect:  fetchBasicTapped,
			},
			Typecycling("Halflingcycling", "{4}", "a Halfling card",
				func(c game.Card) bool { return c.HasSubtypeWord("Halfling") }),
		},
	})
}
