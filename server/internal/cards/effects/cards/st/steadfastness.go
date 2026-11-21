package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfastness — Sorcery {1}{W} (Portal, #5630):
//
//	"Creatures you control get +0/+3 until end of turn."
//
// Warrior's Honor's body with +0/+3 (creaturesYouControlGetUntilEOT):
// the creatures you control as it resolves (CR 611.2c), pumped until
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f8dfc24-f466-4607-8991-acbcbb415db3",
		Name:         "Steadfastness",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.CreaturesYouControlGetUntilEOT(0, 3, "Steadfastness — creatures you control get +0/+3"),
	})
}
