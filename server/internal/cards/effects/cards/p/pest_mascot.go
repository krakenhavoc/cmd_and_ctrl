package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b8ecddf7-ba89-40dd-9f63-5bad27dd850d",
		Name:            "Pest Mascot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Pest Mascot — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
