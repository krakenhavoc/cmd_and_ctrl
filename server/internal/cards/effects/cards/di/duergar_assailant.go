package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duergar Assailant — Creature — Dwarf Soldier {R/W}, 1/1 (Eventide, S82 OLD
// batch 520, #5590):
//
//	"Sacrifice this creature: It deals 1 damage to target attacking or blocking
//	 creature."
//
// Balduvian Trading Post's single-target damage with the sacrifice as the cost,
// paid before the ability goes on the stack, so the Assailant deals the damage by
// its last-known information (CR 608.2h). The target is attacking or blocking as
// it is chosen and again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85b52f70-b583-4f04-a2ab-226d7291c03a",
		Name:         "Duergar Assailant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: It deals 1 damage to target attacking or blocking creature.",
			Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
