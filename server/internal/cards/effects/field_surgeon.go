package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Field Surgeon — Creature — Human Cleric {1}{W}, 1/1 (Urza's Destiny,
// #5621):
//
//	"Tap an untapped creature you control: Prevent the next 1 damage that
//	 would be dealt to target creature this turn."
//
// Opposition's cost (udsB2TapAnUntappedCreatureYouControl) with
// Caregiver's one-damage shield (CR 615.7). Tapping a creature is not
// the {T} symbol (CR 302.6), so the Surgeon can pay with itself the
// turn it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "730809d7-cb8c-495c-8972-7e85b967358e",
		Name:         "Field Surgeon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap an untapped creature you control: Prevent the next 1 damage that would be dealt to target creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    udsB2TapAnUntappedCreatureYouControl(),
			Targets: TargetCreature("target creature"),
			Effect:  preventNextOneDamageToTheTarget("Field Surgeon"),
		}},
	})
}
