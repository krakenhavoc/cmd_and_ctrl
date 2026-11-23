package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spitting Sliver — Creature — Sliver {4}{B}, 3/3 (Planar Chaos, #5595):
//
//	"All Sliver creatures have first strike."
//
// A layer-6 keyword grant (CR 613.1f) of first strike to every Sliver creature at
// the table, an opponent's and this one included (Fury Sliver's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8415a482-3957-4c00-9015-99defc049877",
		Name:         "Spitting Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "first strike"),
		},
	})
}
