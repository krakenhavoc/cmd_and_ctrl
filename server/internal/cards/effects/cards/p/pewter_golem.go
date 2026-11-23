package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pewter Golem — Artifact Creature — Golem {5}, 4/2 (Mirrodin, #5606):
//
//	"{1}{B}: Regenerate this creature."
//
// The shared regenerate row (regenerate_this.go): a shield on the
// ability's source, used the next time it would be destroyed this
// turn (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9d2c69e-6a4a-45b0-aff1-c19bcbe3b490",
		Name:         "Pewter Golem",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}")},
	})
}
