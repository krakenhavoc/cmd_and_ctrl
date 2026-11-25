package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horizon Chimera — Creature — Chimera {2}{G}{U}, 3/2 (Commander 2016, #5537):
//
//	"Flash
//	 Flying, trample
//	 Whenever you draw a card, you gain 1 life."
//
// One trigger per card drawn (WheneverYouDraw), each gaining 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a1ed77f9-51ac-4cca-a9ce-d3874f3948ad",
		Name:            "Horizon Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Horizon Chimera — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
