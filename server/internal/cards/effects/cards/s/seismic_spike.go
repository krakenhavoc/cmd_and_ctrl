package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seismic Spike — Sorcery {2}{R}{R} (Ravnica: City of Guilds, S82 OLD batch
// 581, #5600):
//
//	"Destroy target land. Add {R}{R}."
//
// Two sentences as steps of one walk (CR 608.2c): the destroy waits for any
// prompt its replacement window asks before the mana is added. With the land
// an illegal target as the spell resolves, the spell is countered on
// resolution (CR 608.2b) and adds nothing; a land that is indestructible
// still leaves the mana, since the second sentence does not depend on the
// first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25ec6309-0ccb-4f27-80f1-ec6dcc6461fa",
		Name:         "Seismic Spike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.KineticOozeDestroy), effects.AddMana{Produced: "{R}{R}"}),
	})
}
