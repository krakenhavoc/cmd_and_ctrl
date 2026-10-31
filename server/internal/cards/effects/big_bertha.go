package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Big Bertha — Legendary Creature — Mutant Hero {X}{G}, 1/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"Vigilance
//	 Big Bertha enters with X +1/+1 counters on her.
//	 {1}{G}, {T}: Put a +1/+1 counter on Big Bertha."
//
// Mikaeus, the Lunarch's X counters and its put-a-counter-on-self body.
func init() {
	Register(Spec{
		OracleID:                   "32654ffa-b869-4f8b-a153-4242e8f9d92e",
		Name:                       "Big Bertha",
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, {T}: Put a +1/+1 counter on Big Bertha.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}{G}"), TapCost()),
			Effect:  putCounterOnSelf,
		}},
	})
}
