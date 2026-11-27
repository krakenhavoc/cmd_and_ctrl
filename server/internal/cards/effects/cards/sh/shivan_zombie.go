package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shivan Zombie — Creature — Phyrexian Barbarian Zombie {B}{R}, 2/2 (Invasion, S82 #5616):
//
//	"Protection from white"
//
// Protection from white is a printed keyword (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb3d438e-5e04-45a0-8d16-db63f0faff75",
		Name:            "Shivan Zombie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
	})
}
