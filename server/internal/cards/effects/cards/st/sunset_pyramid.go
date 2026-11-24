package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunset Pyramid — Artifact {2} (Commander Legends):
//
//	"This artifact enters with three brick counters on it.
//	 {2}, {T}, Remove a brick counter from this artifact: Draw a card.
//	 {2}, {T}: Scry 1."
//
// Ring of Three Wishes' counters and cost; the scry ability costs no
// counter, so it works after the bricks run out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5972f426-4a51-4130-baf5-25407c1499f1",
		Name:         "Sunset Pyramid",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("brick", 3, "Sunset Pyramid: enters with three brick counters"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}, Remove a brick counter from this artifact: Draw a card.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.RemoveCountersFromThis("brick", 1)),
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{2}, {T}: Scry 1.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Purpose: game.Purpose{Scry: 1, Answers: game.AnswerValue},
				Effect:  effects.Do(effects.Scry{N: 1}),
			},
		},
	})
}
