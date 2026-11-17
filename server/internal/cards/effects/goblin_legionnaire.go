package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "582fdfbe-c5fe-45d0-b3ff-e66db5183725",
		Name:         "Goblin Legionnaire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}, Sacrifice this creature: It deals 2 damage to any target.",
				Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 2)),
				Effect:  b33DamageChosenTargetFromSource(2),
			},
			{
				Label:   "{W}, Sacrifice this creature: Prevent the next 2 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    Plus(ManaCost("{W}"), SacrificeThis()),
				Targets: TargetAny(),
				Effect:  preventNextDamageToTheTarget("Goblin Legionnaire", 2),
			},
		},
	})
}
