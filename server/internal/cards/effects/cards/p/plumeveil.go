package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plumeveil — Creature — Elemental {W/U}{W/U}{W/U}, 4/4 (Shadowmoor,
// #5591):
//
//	"Flash
//	 Defender, flying"
//
// Flash, defender and flying are printed keywords, as on Angelic Wall and
// Benalish Knight.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "14fa0771-cf2b-4563-98b2-ff4ee24bce21",
		Name:            "Plumeveil",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "defender", "flying"},
	})
}
