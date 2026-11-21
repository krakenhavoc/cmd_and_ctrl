package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Refugee — Creature — Human Warrior {1}{R}{R}, 3/2 (New Phyrexia, #5575):
//
//	"Protection from red"
//
// Repentant Blacksmith's keyword (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7220769a-9d63-4211-9365-9feaf4d9a518",
		Name:            "Vulshok Refugee",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
	})
}
