package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandbar Serpent — Creature — Serpent 3/4 {4}{U} (Urza's Saga, S82 OLD b722):
//
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Cycling is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30fce201-03a8-4afc-a1ad-7a9b3fd59836",
		Name:         "Sandbar Serpent",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
