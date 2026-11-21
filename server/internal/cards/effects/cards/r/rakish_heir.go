package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakish Heir — Creature — Vampire {2}{R}, 2/2 (Innistrad: Midnight Hunt
// Commander reprint):
//
//	"Whenever a Vampire you control deals combat damage to a player, put a
//	 +1/+1 counter on it."
//
// One trigger per Vampire that connects, the counter on that Vampire (a
// creature that has left does nothing, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3318c391-7aa1-4550-88a0-b29d8dca28b7",
		Name:         "Rakish Heir",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82Rpr29VampireCounterTrigger("Rakish Heir")},
	})
}
