package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bilious Skulldweller — Creature — Phyrexian Insect {B}, 1/1 (Phyrexia: All Will Be One, #5479):
//
//	"Deathtouch
//	 Toxic 1"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83bf5610-a798-4508-88d2-c60d96f60d23",
		Name:            "Bilious Skulldweller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "toxic 1"},
	})
}
