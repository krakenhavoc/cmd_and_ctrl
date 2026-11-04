package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gristle Glutton — Creature — Goblin Scout {1}{R}, 1/3:
//
//	"{T}, Blight 1: Discard a card. If you do, draw a card. (To blight 1,
//	 put a -1/-1 counter on a creature you control.)"
//
// The blight is part of the cost (BlightN, #4684): the activator names
// any creature they control, the Glutton itself included, as they
// activate, and the counter goes on as the cost is paid, so Doubling
// Season doesn't double it (CR 614.16). With no creature to blight the
// ability can't be activated (CR 701.68b). The discard is chosen as the
// ability resolves; with an empty hand nothing is discarded and nothing
// is drawn ("if you do").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b7163800-41fc-410a-9650-50ae90842170",
		Name:         "Gristle Glutton",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Blight 1: Discard a card. If you do, draw a card.",
			Cost:    Plus(TapCost(), BlightN(1)),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Effect:  discardACardIfYouDoDraw,
		}},
	})
}
