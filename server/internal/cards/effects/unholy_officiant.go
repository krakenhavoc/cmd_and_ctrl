package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unholy Officiant — Creature — Vampire Cleric {W}, 1/2 (Innistrad:
// Crimson Vow):
//
//	"Vigilance
//	 {4}{W}: Put a +1/+1 counter on this creature."
//
// Adaptive Gemguard's counter body behind a mana cost, at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "41c06726-0360-4488-8dc3-ec257386a372",
		Name:            "Unholy Officiant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{4}{W}"),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
