package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Legionnaire — Creature — Goblin Soldier {R}{W}, 2/2 (Apocalypse,
// #5614):
//
//	"{R}, Sacrifice this creature: It deals 2 damage to any target.
//	 {W}, Sacrifice this creature: Prevent the next 2 damage that would be
//	 dealt to any target this turn."
//
// Two sacrifice abilities, each with the sacrifice as part of its cost,
// so only one of them can ever be activated. The damage is dealt by the
// Legionnaire as it last existed on the battlefield (CR 608.2h), Goblin
// Firestarter's body at 2. The shield is Kitsune Healer's charged one at
// 2 (CR 615.7): it lasts until end of turn and is used up as it
// prevents. A target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "582fdfbe-c5fe-45d0-b3ff-e66db5183725",
		Name:         "Goblin Legionnaire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, Sacrifice this creature: It deals 2 damage to any target.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.B33DamageChosenTargetFromSource(2),
			},
			{
				Label:   "{W}, Sacrifice this creature: Prevent the next 2 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextDamageToTheTarget("Goblin Legionnaire", 2),
			},
		},
	})
}
