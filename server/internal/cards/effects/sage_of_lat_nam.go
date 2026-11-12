package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of Lat-Nam — Creature — Human Artificer {1}{U}, 1/2
// (Antiquities, #5640):
//
//	"{T}, Sacrifice an artifact: Draw a card."
//
// Trading Post's last row without its {1}: a tap and the sacrifice of
// one artifact the controller picks at announce, both paid before the
// ability goes on the stack, then a draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f34be3cc-ff47-4415-a6a8-ed142891dc0c",
		Name:         "Sage of Lat-Nam",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), SacrificeN(1, "an artifact", Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
