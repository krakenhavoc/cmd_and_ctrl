package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siege Zombie — Creature — Zombie {1}{B}, 2/2 (Midnight Hunt, #5497):
//
//	"Tap three untapped creatures you control: Each opponent loses 1 life."
//
// The cost is a TapOthers component, not the {T} symbol (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7756bc2-26ce-4044-a322-a898af8a393d",
		Name:         "Siege Zombie",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap three untapped creatures you control: Each opponent loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    s82MidB09TapThreeCreatures(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return EachOpponentLosesLife{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
