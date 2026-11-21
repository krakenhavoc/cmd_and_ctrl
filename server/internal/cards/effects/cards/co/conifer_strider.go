package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conifer Strider — Creature — Elemental {3}{G}, 5/1 (Dragons of Tarkir,
// #5551):
//
//	"Hexproof (This creature can't be the target of spells or abilities your opponents control.)"
//
// Hexproof is an engine keyword carried as printed data (CR 702.11).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6175197f-cf3c-41df-9d24-b4fffa09631a",
		Name:            "Conifer Strider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
