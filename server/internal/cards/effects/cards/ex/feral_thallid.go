package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Thallid — Creature — Fungus {3}{G}{G}{G}, 6/3 (Fallen Empires, S82 OLD
// b853, #5637):
//
//	"At the beginning of your upkeep, put a spore counter on this creature.
//	 Remove three spore counters from this creature: Regenerate this
//	 creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3953fd17-03f4-4431-99b5-87dcac5a279d",
		Name:         "Feral Thallid",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB853SporeUpkeep("Feral Thallid")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three spore counters from this creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.S82OldB853SporeCost(),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
