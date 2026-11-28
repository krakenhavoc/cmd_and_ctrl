package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bayou Dragonfly — Creature — Insect {1}{G}, 1/1 (Tempest, S82 OLD batch 767,
// #5628):
//
//	"Flying; swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)"
//
// Two printed keywords. The two evasions are independent: a defender needs a
// flying or reach blocker, and a Swamp makes it unblockable outright.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eeb58cb4-b3da-400b-8bfb-ee985cd60f4b",
		Name:            "Bayou Dragonfly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "swampwalk"},
	})
}
