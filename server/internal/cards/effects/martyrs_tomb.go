package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martyrs' Tomb — Enchantment {2}{W}{B} (Apocalypse, #5614):
//
//	"Pay 2 life: Prevent the next 1 damage that would be dealt to target
//	 creature this turn."
//
// Benevolent Ancestor's one-point shield (CR 615.7) on a creature, paid
// for in life rather than a tap, so it can be activated as many times as
// you have life to pay (CR 119.4: you can't pay more life than you have).
// Each activation makes a separate shield of 1 that lasts until end of
// turn. A target gone by resolution gets none (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0baa9eb-0b1e-4d90-a7a1-f7102f575467",
		Name:         "Martyrs' Tomb",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Pay 2 life: Prevent the next 1 damage that would be dealt to target creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    PayLife(2),
			Targets: TargetCreature("target creature"),
			Effect:  preventNextOneDamageToTheTarget("Martyrs' Tomb"),
		}},
	})
}
