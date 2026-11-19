package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minotaur Sureshot — Creature — Minotaur Archer {2}{R}, 2/3 (Foundations
// Jumpstart):
//
//	"Reach
//	 {1}{R}: This creature gets +1/+0 until end of turn."
//
// Akroma's firebreathing shape: any number of activations stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cbb6c481-bcf1-459d-a6fb-bb46cc521368",
		Name:            "Minotaur Sureshot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Minotaur Sureshot — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
