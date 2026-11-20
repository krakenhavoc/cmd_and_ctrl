package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carnivorous Moss-Beast — Creature — Plant Elemental Beast {4}{G}{G},
// 4/5 (Magic 2015, #5556):
//
//	"{5}{G}{G}: Put a +1/+1 counter on this creature."
//
// Hungry Ghoul's counter body (putACounterOnThis) behind a mana-only
// cost: no {T}, so it can be activated any number of times, the turn the
// Moss-Beast arrives included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32f5f67b-bba3-427c-ab24-13a50acdc611",
		Name:         "Carnivorous Moss-Beast",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}{G}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{5}{G}{G}"),
			Effect:  putACounterOnThis(game.CounterPlusOne),
		}},
	})
}
