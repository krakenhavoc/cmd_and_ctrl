package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Larder Zombie — Creature — Zombie {U}, 1/3 (Midnight Hunt, #5497):
//
//	"Defender
//	 Tap three untapped creatures you control: Surveil 1."
//
// The cost is a TapOthers component, not the {T} symbol (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61ab4548-5da9-4351-87e7-a3afc93d4bf5",
		Name:            "Larder Zombie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "Tap three untapped creatures you control: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    s82MidB09TapThreeCreatures(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
