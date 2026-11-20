package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Lashclaw — Creature — Lizard Warrior {1}{R}, 2/2 (Modern
// Horizons 2, #5501):
//
//	"{T}, Discard a card: Creatures you control gain haste until end of
//	 turn."
//
// The {T} waits out summoning sickness (CR 302.6); the discard is a cost
// named at announce. The creatures are chosen once as the ability
// resolves (CR 611.2c), so one that arrives later doesn't gain haste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02358f98-99d5-45a3-b7e3-f2cbf08fa52c",
		Name:         "Viashino Lashclaw",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard a card: Creatures you control gain haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(TapCost(), DiscardN(1, "a card")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"haste"},
					Label:    "Viashino Lashclaw — creatures you control gain haste",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
