package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taoist Hermit — Creature — Human Mystic {2}{G}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Hexproof (This creature can't be the target of spells or abilities your opponents control.)"
//
// Hexproof is an engine keyword carried as printed data: it can't be the
// target of spells or abilities its controller's opponents control
// (CR 702.11).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "960b1251-12af-4766-8df1-bb98e7e967f5",
		Name:            "Taoist Hermit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
