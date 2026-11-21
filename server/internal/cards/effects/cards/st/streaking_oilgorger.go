package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Streaking Oilgorger — Creature — Vampire {4}{B}, 3/3:
//
//	"Flying, haste
//	 Start your engines!
//	 Max speed — This creature has lifelink."
//
// ADR 0138 (#2711).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c163a95e-9414-4b8c-b5a2-1e7674530404",
		Name:            "Streaking Oilgorger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste", effects.StartYourEngines},
		Static:          effects.MaxSpeedSelfKeywords("lifelink"),
	})
}
