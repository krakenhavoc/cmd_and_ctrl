package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadmalkin — Creature — Zombie Cat {B}, 1/1 (Ravnica Remastered, #5948):
//
//	"Menace
//	 {2}{B}, Sacrifice another creature or planeswalker: Put two +1/+1
//	 counters on this creature."
//
// "Another" is object identity (SacrificeAnotherN).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "118cf682-a532-42e9-8d91-6ce6f0d31ace",
		Name:            "Dreadmalkin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice another creature or planeswalker: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeAnotherN(1, "another creature or planeswalker", Or(Creature(), Planeswalker()))),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
