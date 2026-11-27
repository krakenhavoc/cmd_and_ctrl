package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heart Sliver — Creature — Sliver {1}{R}, 1/1 (Tempest, S82 OLD, #5628):
//
//	"All Sliver creatures have haste."
//
// Firewake Sliver's grant, one layer-6 keyword over every Sliver creature at the
// table (CR 613.1f), this one and an opponent's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dd1f375-e8f2-4725-8fbd-9750c4860820",
		Name:         "Heart Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "haste"),
		},
	})
}
