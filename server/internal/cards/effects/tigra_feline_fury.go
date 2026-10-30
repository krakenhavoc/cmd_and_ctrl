package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tigra, Feline Fury — {1}{G} Legendary Creature — Cat Human Hero 2/1
// (Marvel Super Heroes, #4455):
//
//	"Flash
//	 Trample
//	 Whenever you gain life, put a +1/+1 counter on Tigra."
//
// WheneverYouGainLife fires once per gain event, not once per point
// (Ajani's Pridemate), so a 3-life gain is one counter, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39d993be-30ad-4756-a494-fd8808d474c9",
		Name:            "Tigra, Feline Fury",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Tigra — put a +1/+1 counter on her", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
