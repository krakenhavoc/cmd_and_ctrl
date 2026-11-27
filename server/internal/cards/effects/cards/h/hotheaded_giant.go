package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hotheaded Giant — {3}{R} Creature — Giant Warrior, 4/4:
//
//	"Haste
//	 This creature enters with two -1/-1 counters on it unless you've cast
//	 another red spell this turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f90b2cc6-9edb-4666-849a-7c1f6d669daa",
		Name:            "Hotheaded Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements: []game.ReplacementEffect{
			effects.S82OldB522EntersWithMinusCountersUnlessCastAnotherSpellOfColor("R", 2, "red"),
		},
	})
}
