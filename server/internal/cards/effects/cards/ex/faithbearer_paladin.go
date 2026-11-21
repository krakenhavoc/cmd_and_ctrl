package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faithbearer Paladin — Creature — Human Knight {4}{W}, 3/4 (Eldritch
// Moon, #5541):
//
//	"Lifelink"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "725dc451-cae4-4345-9741-6c58e5a1a1dc",
		Name:            "Faithbearer Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
