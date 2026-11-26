package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Hulk — Artifact Creature — Phyrexian Golem {6}, 5/4 (Magic 2013,
// S82 OLD, #5568): no abilities. A vanilla body, registered so the card is in
// the catalog and read as complete.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a16509da-ad46-4ee3-86f3-06c521c23481",
		Name:         "Phyrexian Hulk",
		Completeness: effects.CompletenessFull,
	})
}
