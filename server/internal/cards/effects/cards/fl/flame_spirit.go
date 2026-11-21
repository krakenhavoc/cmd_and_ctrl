package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Spirit — Creature — Elemental Spirit {4}{R}, 2/3 (Ice Age,
// #5636):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing, the shared selfPumpRow at +1/+0 for {R}: no tap, so
// any number of activations, summoning sickness or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "123bfa86-4265-448f-97cd-2c0614212862",
		Name:         "Flame Spirit",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Flame Spirit", "{R}", 1, 0)},
	})
}
