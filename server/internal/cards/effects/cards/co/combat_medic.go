package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Combat Medic — Creature — Human Cleric Soldier {2}{W}, 0/2 (Fallen Empires, S82
// OLD batch 852, #5637):
//
//	"{1}{W}: Prevent the next 1 damage that would be dealt to any target this
//	 turn."
//
// Abuna Acolyte's charged prevention shield (CR 615.7) behind a mana cost instead
// of {T}, so it can be activated any number of times and before it can attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "541b7b2f-7bb0-4858-8dc5-6e50de443c33",
		Name:         "Combat Medic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Combat Medic"),
		}},
	})
}
