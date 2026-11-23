package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lorescale Coatl — Creature — Snake {1}{G}{U}, 2/2 (Alara Reborn,
// #5586):
//
//	"Whenever you draw a card, put a +1/+1 counter on this creature."
//
// Clinquant Skymage's trigger: one counter per card its controller draws,
// each draw its own trigger (CR 603.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c87440aa-4616-49ff-a211-ebdc37767f62",
		Name:         "Lorescale Coatl",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Lorescale Coatl — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
