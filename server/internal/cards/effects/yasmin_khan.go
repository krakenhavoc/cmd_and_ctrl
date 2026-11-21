package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yasmin Khan — Legendary Creature — Human Detective {3}{R}, 3/3 (Doctor Who):
//
//	"{T}: Exile the top card of your library. Until your next end step, you may
//	 play it.
//	 Doctor's companion (You can have two commanders if the other is the Doctor.)"
//
// The window is Dragonhawk's: the permission to play the card (a land too) lasts
// until the end step of the controller's next turn that is at or after now, i.e. the
// next end step of any turn of theirs (game.UntilYourNextEndStepDuration). The card
// stays in exile if it is not played. Doctor's companion is a deck-construction rule
// read off the oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93637599-d9b4-4f7e-a5c0-5a92656e941e",
		Name:         "Yasmin Khan",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Exile the top card of your library. Until your next end step, you may play it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				d := g.UntilYourNextEndStepDuration(item.Controller)
				return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1, Window: &d}.Apply(NewContext(g, item))
			},
		}},
	})
}
