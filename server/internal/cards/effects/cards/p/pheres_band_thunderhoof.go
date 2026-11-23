package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pheres-Band Thunderhoof — Creature — Centaur Warrior {4}{G}, 3/4 (Journey into Nyx, #5558):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put
//	 two +1/+1 counters on this creature."
//
// Satyr Hoplite's heroic with two counters. A Thunderhoof that left and
// came back is a new object and gets none (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "651a1f3d-d6f0-4a38-a776-bceba698127b",
		Name:         "Pheres-Band Thunderhoof",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Pheres-Band Thunderhoof — two +1/+1 counters on it (heroic)", effects.ThisGetsTwoPlusOneCounters),
		},
	})
}
