package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodthirsty Aerialist — Creature — Vampire Rogue {1}{B}{B}, 2/3
// (Foundations Jumpstart):
//
//	"Flying
//	 Whenever you gain life, put a +1/+1 counter on this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae78b6f9-8334-479e-8111-8e3acb2b7ca7",
		Name:            "Bloodthirsty Aerialist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.LifeGainPaying(
				effects.WheneverYouGainLife("Bloodthirsty Aerialist — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
	})
}
