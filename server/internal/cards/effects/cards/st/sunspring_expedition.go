package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunspring Expedition — Enchantment {W} (Zendikar, S82 OLD b496):
//
//	"Landfall — Whenever a land you control enters, you may put a quest counter on
//	 this enchantment.
//	 Remove three quest counters from this enchantment and sacrifice it: You gain 8
//	 life."
//
// Khalni Heart Expedition's landfall counter and cost with a fixed life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5aeed95-24f8-4f28-9e7c-1e3502d013cd",
		Name:         "Sunspring Expedition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Sunspring Expedition — put a quest counter on it",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Put a quest counter on Sunspring Expedition?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three quest counters from this enchantment and sacrifice it: You gain 8 life.",
			Purpose: game.Purpose{LifeGain: 8, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 3), effects.SacrificeThis()),
			Effect:  effects.Do(effects.GainLife{Amount: 8}),
		}},
	})
}
