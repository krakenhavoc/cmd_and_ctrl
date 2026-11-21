package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cast Out — Enchantment {3}{W}:
//
//	"Flash
//	 When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield.
//	 Cycling {W}"
//
// Banishing Light's two rows (CR 610.3) with flash and the shared cycling
// constructor. If Cast Out leaves before the trigger resolves nothing is
// exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f90b00f6-36e0-4988-9409-57297483a952",
		Name:            "Cast Out",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       effects.ExileUntilThisLeavesTriggers("Cast Out"),
		Activated:       []effects.ActivatedAbility{effects.Cycling("{W}")},
	})
}
