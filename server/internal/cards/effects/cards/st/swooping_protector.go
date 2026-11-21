package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swooping Protector — Creature — Bird Citizen {3}{W}, 2/1 (Streets of
// New Capenna):
//
//	"Flash
//	 Flying
//	 This creature enters with a shield counter on it. (If it would be
//	 dealt damage or destroyed, remove a shield counter from it
//	 instead.)"
//
// The counter is a self-entry replacement, so it is on the creature as
// it enters, and CR 122.1c's built-ins do the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "094689cc-617a-426e-beb9-3f8de692cd5b",
		Name:            "Swooping Protector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Replacements:    effects.Snc6Entry("Swooping Protector"),
	})
}
