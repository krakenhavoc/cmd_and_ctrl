package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borrowed Time — Enchantment {2}{W} (Innistrad: Midnight Hunt):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Banishing Light's text word for word (CR 610.3): the shared
// exileUntilThisLeavesTriggers rows. If the enchantment leaves before the
// trigger resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1ac4e0e-5633-46e4-9a21-d3924e481d13",
		Name:         "Borrowed Time",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.ExileUntilThisLeavesTriggers("Borrowed Time"),
	})
}
