package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death-Hood Cobra — Creature — Phyrexian Snake {1}{G}, 2/2 (New Phyrexia, #5575):
//
//	"{1}{G}: This creature gains reach until end of turn.
//	 {1}{G}: This creature gains deathtouch until end of turn."
//
// Two self grants until end of turn (somB1SelfGainsUntilEOT), each its
// own row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "747dfffd-880c-48fe-9380-eb56071f5910",
		Name:         "Death-Hood Cobra",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.SomB1SelfGainsUntilEOT("Death-Hood Cobra", "{1}{G}", "reach"),
			effects.SomB1SelfGainsUntilEOT("Death-Hood Cobra", "{1}{G}", "deathtouch"),
		},
	})
}
