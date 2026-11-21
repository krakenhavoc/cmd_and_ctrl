package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greater Basilisk — Creature — Basilisk {3}{G}{G}, 3/5 (Magic 2011,
// #5579):
//
//	"Deathtouch"
//
// Deathtouch is the printed keyword the state-based check reads (CR
// 704.5h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e42e7dc-acd9-4125-a90d-75fd8da204b3",
		Name:            "Greater Basilisk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
