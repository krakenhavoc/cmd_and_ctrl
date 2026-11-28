package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sawback Manticore — Creature — Manticore {3}{R}{G}, 2/4 (Mirage, #5633):
//
//	"{4}: This creature gains flying until end of turn.
//	 {1}: This creature deals 2 damage to target attacking or blocking creature.
//	 Activate only if this creature is attacking or blocking and only once each
//	 turn."
//
// The window is read at announce, the once-a-turn limit counts announcements
// (CR 602.1b).
//
// No simplification.
func init() {
	const damageLabel = "{1}: This creature deals 2 damage to target attacking or blocking creature. Activate only if this creature is attacking or blocking and only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "9360d30c-4556-4b41-bd50-08b4367e5efb",
		Name:         "Sawback Manticore",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{4}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Sawback Manticore — flying until end of turn"),
			},
			{
				Label:     damageLabel,
				Cost:      effects.ManaCost("{1}"),
				Targets:   effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
				Purpose:   effects.ForTargets(effects.DamageToTarget(0, 2)),
				Condition: effects.AllConditions(effects.OldB794SourceIsAttackingOrBlocking(), effects.OncePerTurnActivation(damageLabel)),
				Effect:    effects.SourceDealsDamageToEachLegalTarget(2),
			},
		},
	})
}
