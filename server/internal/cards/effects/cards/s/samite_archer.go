package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Samite Archer — Creature — Human Cleric Archer {1}{W}{U}, 1/1 (Invasion, S82 #5616):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this turn.
//	 {T}: This creature deals 1 damage to any target."
//
// Samite Healer's prevention shield and Prodigal Sorcerer's ping on one body:
// two separate tap abilities, so only one can be used a turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5322a218-6a0c-4d86-835a-c87eb52aaf6a",
		Name:         "Samite Archer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextOneDamageToTheTarget("Samite Archer"),
			},
			{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			},
		},
	})
}
