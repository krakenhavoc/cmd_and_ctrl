package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pest Mascot — Creature — Pest Ape {1}{B}{G}, 2/3 (Secrets of Strixhaven,
// #4456):
//
//	"Trample
//	 Whenever you gain life, put a +1/+1 counter on this creature."
//
// Ajani's Pridemate's trigger: one gain event is one counter, however much
// life it gained.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b8ecddf7-ba89-40dd-9f63-5bad27dd850d",
		Name:            "Pest Mascot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Pest Mascot — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
