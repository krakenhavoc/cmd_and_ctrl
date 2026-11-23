package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mischievous Sneakling — Creature — Shapeshifter {1}{U/B} (Lorwyn Eclipsed,
// #4458):
//
//	"Changeling (This card is every creature type.)
//	 Flash"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe7a4361-561a-4bab-921e-d01bab19dad3",
		Name:            "Mischievous Sneakling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling", "flash"},
	})
}
