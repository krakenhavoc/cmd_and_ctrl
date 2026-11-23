package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Douse in Gloom — Instant {2}{B} (Guildpact, #5599):
//
//	"Douse in Gloom deals 2 damage to target creature and you gain 2
//	 life."
//
// Vicious Hunger's body at instant speed: damageTheClauseTarget, then
// the life. If the creature is an illegal target as the spell resolves,
// the spell does nothing at all (CR 608.2b), so no life is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1fd754fd-c053-4638-b379-46d903357b01",
		Name:         "Douse in Gloom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), LifeGain: 2, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 2), effects.GainLife{Amount: 2}),
	})
}
