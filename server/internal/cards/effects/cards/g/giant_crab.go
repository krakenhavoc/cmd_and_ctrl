package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Crab — Creature — Crab {4}{U}, 3/3 (Tempest, S82 OLD b753, #5628):
//
//	"{U}: This creature gains shroud until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "780f029f-e2e1-431b-9a8b-a1af51d7da77",
		Name:         "Giant Crab",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ConB1ThisGainsRow("{U}", "shroud", "Giant Crab", game.AnswerProtect)},
	})
}
