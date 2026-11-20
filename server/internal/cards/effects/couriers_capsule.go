package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Courier's Capsule — Artifact {1}{U} (Shards of Alara, #5588):
//
//	"{1}{U}, {T}, Sacrifice this artifact: Draw two cards."
//
// Hedron Archive's draw on a cheaper Capsule: the mana, the tap and the
// sacrifice are all costs (CR 602.2b), and the draw resolves for the
// activator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e694613-e2e1-4ae3-baa8-f372259e91cf",
		Name:         "Courier's Capsule",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Cost:    Plus(ManaCost("{1}{U}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
