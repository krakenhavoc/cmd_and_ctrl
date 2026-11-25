package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonds of Quicksilver — Enchantment — Aura {3}{U} (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Bitter Chill's untap lock (EnchantedDoesntUntap) without its tap: the
// creature stays tapped only if it already was, and an untapped one stays
// untapped until something taps it. Flash is the engine's keyword
// (CR 702.8), so the Aura may be cast in response, even after attackers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "03ce87f1-cdda-4121-8c0a-988642053055",
		Name:                  "Bonds of Quicksilver",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
