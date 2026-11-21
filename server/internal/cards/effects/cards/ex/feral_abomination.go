package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Abomination — Creature — Thrull {5}{B}, 5/5 (Dominaria, #5530):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)"
//
// One printed keyword, read by the state-based check after it deals damage (CR 702.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1bf06342-0d76-433b-9912-8af03b7e3bc3",
		Name:            "Feral Abomination",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
