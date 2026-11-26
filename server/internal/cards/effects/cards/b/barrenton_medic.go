package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrenton Medic — Creature — Kithkin Cleric {4}{W}, 0/4 (Shadowmoor, S82 OLD
// b532, #5591):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this turn.
//	 Put a -1/-1 counter on this creature: Untap this creature."
//
// Jhessian Balmgiver's one-damage prevention shield on any target, and the
// counter-for-untap that lets the Medic do it again. The Medic dies once its
// -1/-1 counters reach its toughness (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17661062-928c-4e33-a4f3-92ddc9d8c817",
		Name:         "Barrenton Medic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextOneDamageToTheTarget("Barrenton Medic"),
			},
			effects.OldB532UntapThisForAMinusOneCounter(),
		},
	})
}
