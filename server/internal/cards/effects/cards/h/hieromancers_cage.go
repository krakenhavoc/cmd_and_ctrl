package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hieromancer's Cage — Enchantment {3}{W} (Core Set 2019, #5527):
//
//	"When this enchantment enters, exile target nonland permanent an opponent controls until this enchantment leaves the battlefield."
//
// Banishing Light's text: the enters trigger exiles the target until
// this enchantment leaves (CR 610.3), through the shared
// exileUntilThisLeavesTriggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "207c7f93-3abf-44c5-ace6-3f86359c9745",
		Name:         "Hieromancer's Cage",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.ExileUntilThisLeavesTriggers("Hieromancer's Cage"),
	})
}
