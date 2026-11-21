package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodland Patrol — Creature — Human Scout {2}{G}, 3/2 (Eldritch Moon, #5541):
//
//	"Vigilance"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce31cfc4-a109-4cf5-b8f9-210267a98114",
		Name:            "Woodland Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
