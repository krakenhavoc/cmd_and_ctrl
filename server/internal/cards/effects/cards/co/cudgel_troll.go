package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cudgel Troll — Creature — Troll {2}{G}{G}, 4/3 (Magic 2010, #5585):
//
//	"{G}: Regenerate this creature. (The next time this creature would be
//	 destroyed this turn, instead tap it, remove it from combat, and heal
//	 all damage on it.)"
//
// Ancient Silverback's regeneration: each activation puts one shield on the
// Troll (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2f86723-4765-4dac-a9f6-9cfa543613b2",
		Name:         "Cudgel Troll",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{G}")},
	})
}
