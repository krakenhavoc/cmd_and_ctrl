package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimir Spybug — Creature — Insect {U}{B}, 1/1:
//
//	"Flying
//	 Menace
//	 Whenever you surveil, put a +1/+1 counter on this creature."
//
// The payoff is the batch's one surveil trigger (EventSurveil, once a surveil
// is complete) with a counter on the source, which does nothing if the Spybug
// has left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3bb7041c-4f3e-4123-9181-3442ea0c98e8",
		Name:            "Dimir Spybug",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
		Triggered: []game.TriggeredAbility{
			cmdr2B27WheneverYouSurveil("Dimir Spybug — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
