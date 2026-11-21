package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malachite Golem — Artifact Creature — Golem {6}, 5/3 (Mirrodin, #5606):
//
//	"{1}{G}: This creature gains trample until end of turn."
//
// A layer 6 grant pinned to the source until cleanup (CR 514.2),
// the shape Endling's keyword rows use.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54e71732-6575-48f4-838b-35d44eba51cf",
		Name:         "Malachite Golem",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.UdsB3ThisGainsRow("{1}{G}", "trample", "Malachite Golem")},
	})
}
