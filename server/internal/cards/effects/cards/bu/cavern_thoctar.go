package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cavern Thoctar — Creature — Beast {5}{G}, 5/5 (Shards of Alara, S82 OLD
// batch 511, #5588):
//
//	"{1}{R}: This creature gets +1/+0 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a1b8db2-8814-45b3-ab8a-a1c8522d7fa1",
		Name:         "Cavern Thoctar",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{1}{R}", 1, 0, "Cavern Thoctar")},
	})
}
