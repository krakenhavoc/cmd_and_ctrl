package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarlet Witch, Wanda Maximoff — Legendary Creature — Mutant Warlock
// Hero {2}{R}, 2/3 (Marvel Super Heroes Commander, #4455): "Menace". A
// keyword-only body, the same shape as Glamorous Grapplers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "488591d4-3dfb-4e1e-b0c4-1950ff92b848",
		Name:            "Scarlet Witch, Wanda Maximoff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
