package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "39d993be-30ad-4756-a494-fd8808d474c9",
		Name:            "Tigra, Feline Fury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Tigra — put a +1/+1 counter on her", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
