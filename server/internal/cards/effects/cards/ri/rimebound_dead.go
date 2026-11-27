package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimebound Dead — Snow Creature — Skeleton {B}, 1/1 (Coldsnap):
//
//	"{S}: Regenerate this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfc58887-859d-4ba5-8507-1fb1aa4d37c0",
		Name:         "Rimebound Dead",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{S}")},
	})
}
