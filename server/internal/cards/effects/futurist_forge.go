package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Futurist Forge — {1}{U} Artifact:
//
//	"When this artifact enters, draw a card.
//	 {3}{U}, Sacrifice this artifact: Draw two cards."
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:     "0e5f97b5-bb8c-4688-a141-5db4d4eeccf0",
		Name:         "Futurist Forge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Futurist Forge — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{U}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
