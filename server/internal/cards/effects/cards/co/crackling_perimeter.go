package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crackling Perimeter — Enchantment {1}{R} (Gatecrash, S82 OLD #5565):
//
//	"Tap an untapped Gate you control: This enchantment deals 1 damage to
//	 each opponent."
//
// Heap Gate's cost: tapping a Gate is not the {T} symbol (CR 302.6), so a Gate
// that arrived this turn may pay it, and it does not target. The ability has
// no other cost, so it can be activated once per untapped Gate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "413e880a-bdc8-41bc-923a-566b08e68d0b",
		Name:         "Crackling Perimeter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped Gate you control: This enchantment deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapAnUntapped("an untapped Gate you control", effects.Subtype("Gate")),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 1}),
		}},
	})
}
