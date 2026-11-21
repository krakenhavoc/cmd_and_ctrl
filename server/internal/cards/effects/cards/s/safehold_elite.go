package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Safehold Elite — Creature — Elf Scout {1}{G/W}, 2/2 (Shadowmoor, #5591):
//
//	"Persist (When this creature dies, if it had no -1/-1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 -1/-1 counter on it.)"
//
// Persist is a canonical keyword whose dies trigger the engine derives
// from the departed permanent's last-known abilities (CR 702.79a, ADR 0113
// §4, #2075), as on Aerie Ouphes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68ca91ba-31fb-47e0-9b32-e4f3504cbbca",
		Name:            "Safehold Elite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
	})
}
