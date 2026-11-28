package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molting Harpy — Creature — Harpy Mercenary {B}, 2/1 (Mercadian Masques, S82
// OLD b684, #5619):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice this creature unless you pay
//	 {2}."
//
// Flying is a printed keyword; the upkeep payment is the shared "sacrifice
// unless you pay" (Energy Flux, Haazda Shield Mate).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8dcb8d8-ae27-43ac-88e7-e59f2656e0ce",
		Name:            "Molting Harpy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.UpkeepSacrificeUnlessYouPay("Molting Harpy", "{2}")},
	})
}
