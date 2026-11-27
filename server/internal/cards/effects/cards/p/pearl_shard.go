package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pearl Shard — Artifact {3} (Mirrodin, S82 OLD, #5606):
//
//	"{3}, {T} or {W}, {T}: Prevent the next 2 damage that would be dealt to
//	 any target this turn."
//
// Granite Shard's shape: one ability with a choice of cost, written as two
// rows so the activator picks the cost by picking the row. Both resolve as
// Goblin Legionnaire's shield: a charged prevention of 2 on the target
// (CR 615.7), checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	const label = "{3}, {T} or {W}, {T}: Prevent the next 2 damage that would be dealt to any target this turn."
	effects.Register(effects.Spec{
		OracleID:     "d08cb402-2138-4e8e-99f4-4a62855dc285",
		Name:         "Pearl Shard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   label,
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextDamageToTheTarget("Pearl Shard", 2),
			},
			{
				Label:   label,
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextDamageToTheTarget("Pearl Shard", 2),
			},
		},
	})
}
