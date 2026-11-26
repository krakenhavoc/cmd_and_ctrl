package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindlock Orb — Artifact {3}{U} (Shards of Alara, S82 OLD batch 517, #5588):
//
//	"Players can't search libraries."
//
// Stranglehold's search lock over every player, its controller included
// (Spec.PlayersCantSearchLibraries, CR 701.23a), with no way to ignore it.
// Judged through the card's own abilities, so an Orb that has lost its
// abilities blocks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "5794b58e-f3b6-4637-b288-283b3a72a718",
		Name:                       "Mindlock Orb",
		Completeness:               effects.CompletenessFull,
		PlayersCantSearchLibraries: true,
	})
}
