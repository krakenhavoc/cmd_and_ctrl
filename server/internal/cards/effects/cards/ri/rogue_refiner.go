package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rogue Refiner — Creature — Human Rogue {1}{G}{U}, 3/2:
//
//	"When this creature enters, draw a card and you get {E}{E} (two
//	 energy counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2f09e41-0a91-4fb6-8804-874b3a5166b0",
		Name:         "Rogue Refiner",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rogue Refiner — draw a card and you get {E}{E}", effects.Do(effects.DrawCards{N: 1}, effects.GetEnergy{N: 2})),
		},
	})
}
