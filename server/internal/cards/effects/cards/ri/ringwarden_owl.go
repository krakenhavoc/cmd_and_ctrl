package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ringwarden Owl — Creature — Bird {3}{U}{U}, 3/3 (Magic Origins, #5549):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)"
//
// Wing Commando's shape: flying and prowess, both engine keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a8b3316-a01a-4206-bbed-d7aac6c834bd",
		Name:            "Ringwarden Owl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
	})
}
