package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walking Corpse — Creature — Zombie {1}{B}, 2/2 (Innistrad, #5572):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fea95888-e16a-4209-9cd4-623f7f4d2f67",
		Name:         "Walking Corpse",
		Completeness: effects.CompletenessFull,
	})
}
