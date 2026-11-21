package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tattered Drake — Creature — Zombie Drake {4}{U}, 2/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 {B}: Regenerate this creature."
//
// Flying is printed data; the regeneration row is regenerateThisRow (CR
// 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1199a252-3c7c-4ae4-b063-dc132bfbea2f",
		Name:            "Tattered Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}
