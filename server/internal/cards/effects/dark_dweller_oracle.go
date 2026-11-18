package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark-Dweller Oracle — Creature — Goblin Shaman {1}{R}, 2/2:
//
//	"{1}, Sacrifice a creature: Exile the top card of your library. You
//	 may play that card this turn."
//
// The sacrifice is a cost (any creature, itself included); the effect is
// the impulse exile Breeches' bullet shares (exileTopCardYouMayPlayThisTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17917b7a-2b2b-4a96-a21e-10cd91375660",
		Name:         "Dark-Dweller Oracle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Exile the top card of your library. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Impulse: 1},
			Cost:    Plus(ManaCost("{1}"), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTopCardYouMayPlayThisTurn(item, NewContext(g, item), 0)
			},
		}},
	})
}
