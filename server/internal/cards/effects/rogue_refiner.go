package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rogue Refiner — Creature — Human Rogue {1}{G}{U}, 3/2:
//
//	"When this creature enters, draw a card and you get {E}{E} (two
//	 energy counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2f09e41-0a91-4fb6-8804-874b3a5166b0",
		Name:         "Rogue Refiner",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rogue Refiner — draw a card and you get {E}{E}", Do(DrawCards{N: 1}, GetEnergy{N: 2})),
		},
	})
}
