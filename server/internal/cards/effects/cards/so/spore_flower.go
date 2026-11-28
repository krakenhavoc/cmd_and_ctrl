package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spore Flower — Creature — Fungus {G}{G}, 0/1 (Fallen Empires, S82 OLD b853,
// #5637):
//
//	"At the beginning of your upkeep, put a spore counter on this creature.
//	 Remove three spore counters from this creature: Prevent all combat damage
//	 that would be dealt this turn."
//
// Root Snare's prevention, paid with the spores.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f352ec62-bf09-45f2-b5b7-75931177b15f",
		Name:         "Spore Flower",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB853SporeUpkeep("Spore Flower")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three spore counters from this creature: Prevent all combat damage that would be dealt this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.S82OldB853SporeCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PreventAllCombatDamageThisTurn{Label: "Spore Flower: prevent combat damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
