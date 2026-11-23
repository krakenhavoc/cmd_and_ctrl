package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hawkeye, Clint Barton — Legendary Creature — Human Archer Hero
// {3}{W}, 3/5 (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance"
//
// A keyword creature with nothing else on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c9efec35-823f-4c78-be36-d1a57d8b6f44",
		Name:            "Hawkeye, Clint Barton",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
