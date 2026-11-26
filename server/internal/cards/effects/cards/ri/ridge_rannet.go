package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ridge Rannet — Creature — Beast {5}{R}{R}, 6/4 (Shards of Alara):
//
//	"Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A vanilla 6/4 with the shared cycling ability, activated from the hand
// (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "562aaf81-e70a-401d-a638-bb7ce4f40626",
		Name:         "Ridge Rannet",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
