package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Hellhound — Creature — Elemental Dog {1}{R}{R}, 2/2 (Magic 2010, #5585):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// Flame Spirit's firebreathing: a layer 7c pump that ends at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab71c8ba-c7ac-4ad2-a6db-29d903bbca78",
		Name:         "Fiery Hellhound",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Fiery Hellhound", "{R}", 1, 0)},
	})
}
