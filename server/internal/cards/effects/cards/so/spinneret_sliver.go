package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinneret Sliver — Creature — Sliver {1}{G}, 2/2 (Time Spiral, #5596):
//
//	"All Sliver creatures have reach."
//
// A layer-6 keyword grant (CR 613.1f) of reach (CR 702.17a) to every
// Sliver creature at the table, an opponent's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbdec63e-0679-47e9-b82a-daa56df50481",
		Name:         "Spinneret Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "reach"),
		},
	})
}
