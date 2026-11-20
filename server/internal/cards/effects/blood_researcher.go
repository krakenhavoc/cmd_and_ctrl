package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Researcher — Creature — Vampire Druid {1}{B}{G}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Whenever you gain life, put a +1/+1 counter on this creature."
//
// Menace is the printed keyword; the life-gain trigger is Ajani's
// Pridemate's, once per life-gain event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a465ba74-615a-4261-b4a0-fcb0819f10b3",
		Name:            "Blood Researcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			LifeGainPaying(
				WheneverYouGainLife("Blood Researcher — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
	})
}
