package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ant-Man, Scott Lang — Legendary Creature — Human Hero {1}{W}, 2/2
// (Marvel Super Heroes, #4455):
//
//	"{4}: Put a +1/+1 counter on Ant-Man."
//
// Marketback Walker's activated ability on a 2/2: a plain {4} with no
// tap, usable at instant speed any number of times. The counter goes on
// the creature itself through the shared putCounterOnSelf body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e4cf4a4-50a4-4856-bf7e-20899e8affd1",
		Name:         "Ant-Man, Scott Lang",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}: Put a +1/+1 counter on Ant-Man.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{4}"),
			Effect:  putCounterOnSelf,
		}},
	})
}
