package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Smokeweaver — Creature — Bird Soldier {2}{U}{U}, 2/3 (Odyssey, #5613):
//
//	"Flying, protection from red"
//
// Flying (CR 702.9b) and protection from red, which the engine reads off
// the printed keyword: red spells and abilities from red sources can't
// target it (CR 702.16b), damage from red sources is prevented (CR
// 702.16e), and red creatures can't block it (CR 702.16f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e80be353-8b19-42d2-9141-5353a8409c8e",
		Name:            "Aven Smokeweaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
	})
}
