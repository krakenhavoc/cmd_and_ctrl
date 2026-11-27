package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yoked Plowbeast — Creature — Beast {5}{W}{W}, 5/5 (Shards of Alara):
//
//	"Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A vanilla 5/5 with the shared cycling ability, activated from the hand
// (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07237694-b5d5-4c92-a429-55e343b8dd1c",
		Name:         "Yoked Plowbeast",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
