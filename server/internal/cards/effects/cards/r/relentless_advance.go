package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relentless Advance — Sorcery {3}{U} (War of the Spark, #5520):
//
//	"Amass Zombies 3. (Put three +1/+1 counters on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f54573d1-9950-46c3-b8f1-cd6f081d62ad",
		Name:         "Relentless Advance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 3},
		OnResolve:    effects.ResolveSteps(effects.Amass{Subtype: "Zombie", N: 3}),
	})
}
