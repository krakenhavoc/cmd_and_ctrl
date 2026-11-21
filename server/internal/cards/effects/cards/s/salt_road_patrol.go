package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salt Road Patrol — Creature — Human Scout {3}{W}, 2/5 (Khans of
// Tarkir, #5554):
//
//	"Outlast {1}{W} ({1}{W}, {T}: Put a +1/+1 counter on this creature.
//	 Outlast only as a sorcery.)"
//
// Outlast is the shared sorcery-speed tap-and-pay counter row
// (rprB04Outlast, CR 702.107a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b101943b-23f1-450f-a722-13761dc33419",
		Name:         "Salt Road Patrol",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RprB04Outlast("{1}{W}")},
	})
}
