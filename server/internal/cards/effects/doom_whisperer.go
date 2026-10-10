package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doom Whisperer — Creature — Nightmare Demon {3}{B}{B}, 6/6 (EDHREC
// rank 887):
//
//	"Flying, trample
//	 Pay 2 life: Surveil 2."
//
// Flying and trample ride PrintedKeywords. The activated ability is
// an ordinary CR 602 ability with a life-payment cost (PayLife) and
// no mana or tap component, so it can be activated any number of
// times a turn, at instant speed, as many times as the controller
// has life to spare — exactly as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4a01db2e-cd43-4b1a-a480-169018f82501",
		Name:            "Doom Whisperer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Activated: []ActivatedAbility{{
			Label:   "Pay 2 life: Surveil 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    PayLife(2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
