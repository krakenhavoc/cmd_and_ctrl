package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moss Viper — Creature — Snake {G}, 1/1 (Theros Beyond Death, #5514):
//
//	"Deathtouch"
//
// Deathtouch is the printed keyword (Typhoid Rats' shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d6f6ee4-77c2-4986-91c7-7dee962fc1a7",
		Name:            "Moss Viper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
