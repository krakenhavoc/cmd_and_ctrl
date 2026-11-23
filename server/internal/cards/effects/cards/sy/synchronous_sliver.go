package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Synchronous Sliver — Creature — Sliver {4}{U}, 3/3 (Planar Chaos, #5595):
//
//	"All Sliver creatures have vigilance."
//
// A layer-6 keyword grant (CR 613.1f) of vigilance to every Sliver creature at
// the table, an opponent's and this one included (Fury Sliver's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b90e2fb8-9b09-4d95-a8a2-5287b5f18132",
		Name:         "Synchronous Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "vigilance"),
		},
	})
}
