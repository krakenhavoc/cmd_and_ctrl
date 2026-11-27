package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandbar Merfolk — Creature — Merfolk 1/1 {U} (Urza's Saga, S82 OLD b722):
//
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Cycling is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d6e5a88-a7fd-466c-a9f8-5015e3400176",
		Name:         "Sandbar Merfolk",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
