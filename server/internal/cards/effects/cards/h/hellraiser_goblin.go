package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellraiser Goblin — Creature — Goblin Berserker {2}{R}, 2/2 (Duel Decks:
// Speed vs. Cunning, #5555):
//
//	"Creatures you control have haste and attack each combat if able."
//
// Cyclops of Eternal Fury's haste grant and Goblin Rabblemaster's attack
// requirement, both over every creature its controller controls, itself
// included.
//
// No simplification.
func init() {
	mine := effects.TribeFilter{YoursOnly: true}
	atk := effects.AttacksEachCombatWhere(mine.Matches)
	atk.Label = "Creatures you control attack each combat if able."
	effects.Register(effects.Spec{
		OracleID:     "c30e89fa-8993-468d-b2ee-093c3950f520",
		Name:         "Hellraiser Goblin",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(mine, "haste"),
			atk,
		},
	})
}
