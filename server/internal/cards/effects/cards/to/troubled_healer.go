package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Troubled Healer — Creature — Human Cleric {2}{W}, 1/2 (Prophecy):
//
//	"Sacrifice a land: Prevent the next 2 damage that would be dealt to any
//	 target this turn."
//
// Pearl Shard's prevention shield (a charged prevention of 2 on the target,
// CR 615.7, checked as the ability resolves, CR 608.2b) behind Spitting
// Spider's cost: the sacrifice of any land its controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4377cc84-19d3-48e0-9588-d18b4fbf4e6d",
		Name:         "Troubled Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: Prevent the next 2 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.B08SacrificeALand(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextDamageToTheTarget("Troubled Healer", 2),
		}},
	})
}
