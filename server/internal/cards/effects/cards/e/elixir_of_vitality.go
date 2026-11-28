package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elixir of Vitality — Artifact {4} (Mirage, S82 OLD batch 809, #5633):
//
//	"This artifact enters tapped.
//	 {T}, Sacrifice this artifact: You gain 4 life.
//	 {8}, {T}, Sacrifice this artifact: You gain 8 life."
//
// An entry replacement (SelfEntersTapped) and two sacrifice abilities with
// their own costs; both need the artifact untapped ({T} in the cost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "635cc10e-ca25-49a2-af44-3b064263a254",
		Name:         "Elixir of Vitality",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice this artifact: You gain 4 life.",
				Purpose: game.Purpose{LifeGain: 4, Answers: game.AnswerValue},
				Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true},
				Effect:  effects.Do(effects.GainLife{Amount: 4}),
			},
			{
				Label:   "{8}, {T}, Sacrifice this artifact: You gain 8 life.",
				Purpose: game.Purpose{LifeGain: 8, Answers: game.AnswerValue},
				Cost:    game.AbilityCost{Mana: "{8}", Tap: true, SacrificeSelf: true},
				Effect:  effects.Do(effects.GainLife{Amount: 8}),
			},
		},
	})
}
