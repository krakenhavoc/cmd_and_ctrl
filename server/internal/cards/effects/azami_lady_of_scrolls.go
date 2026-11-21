package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azami, Lady of Scrolls — Legendary Creature — Human Wizard {2}{U}{U}{U},
// 0/2 (Commander Masters):
//
//	"Tap an untapped Wizard you control: Draw a card."
//
// The cost is a tap of an untapped Wizard, not the {T} symbol (CR 302.6):
// a Wizard that arrived this turn, Azami herself included, may pay it. It
// does not target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f8b97fe-3e5e-47c2-9a9d-7f77482aa159",
		Name:         "Azami, Lady of Scrolls",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap an untapped Wizard you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapAnUntapped("an untapped Wizard you control", HasSubtype("Wizard")),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
