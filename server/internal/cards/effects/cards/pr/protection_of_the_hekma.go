package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protection of the Hekma — {4}{W} Enchantment (Amonkhet, S82 old batch 248):
//
//	"If a source an opponent controls would deal damage to you, prevent 1 of
//	 that damage."
//
// A CR 615 prevention effect on damage to the controller alone (not to their
// permanents), whose source is read as the damage snapshotted it (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5f611c9-9e6e-4719-ad71-8686327a8564",
		Name:         "Protection of the Hekma",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB248PreventOneFromOpponentsToYou()},
	})
}
