package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serum Sovereign — Creature — Phyrexian Sphinx {4}{U}, 4/4 (Phyrexia: All Will Be One):
//
//	"Flying
//	 Whenever you cast a noncreature spell, put an oil counter on this creature.
//	 {U}, Remove an oil counter from this creature: Draw a card, then scry 2."
//
// The scry waits for the draw (Do runs the steps in order and the scry only
// queues its prompt).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d52ecb32-26ed-484b-b1a7-ee47b66f0f94",
		Name:            "Serum Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.S82OneB09OilOnNoncreatureCast("Serum Sovereign")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Remove an oil counter from this creature: Draw a card, then scry 2.",
			Purpose: game.Purpose{Draws: 1, Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.RemoveCountersFromThis(effects.S82OneB09Oil, 1)),
			Effect:  effects.Do(effects.DrawCards{N: 1}, effects.Scry{N: 2}),
		}},
	})
}
