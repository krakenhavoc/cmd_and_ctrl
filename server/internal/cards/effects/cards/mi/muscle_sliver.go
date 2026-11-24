package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Muscle Sliver — Creature — Sliver {1}{G}, 1/1 (Secret Lair Drop):
//
//	"All Sliver creatures get +1/+1."
//
// A layer-7c bonus for every Sliver creature at the table, this one included,
// whoever controls it (Might Sliver's shape at +1/+1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06e0459f-d691-4470-884a-27c09b6612d2",
		Name:         "Muscle Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}}, 1, 1),
		},
	})
}
