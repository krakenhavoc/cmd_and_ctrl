package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beast Walkers — Creature — Human Beast Soldier {1}{W}{W}, 2/2 (Homelands, S82
// OLD b821, #5635):
//
//	"{G}: This creature gains banding until end of turn."
//
// The creature's own keyword grant (S82OldB510ThisGainsRow); banding is
// ADR 0224's keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68dfc626-66d5-4a40-9410-2b15676ebcec",
		Name:         "Beast Walkers",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{G}", "banding")},
	})
}
