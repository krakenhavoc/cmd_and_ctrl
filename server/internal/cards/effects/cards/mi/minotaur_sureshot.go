package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "cbb6c481-bcf1-459d-a6fb-bb46cc521368",
		Name:            "Minotaur Sureshot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Minotaur Sureshot — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
