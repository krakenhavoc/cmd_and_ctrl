package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fury Sliver — Creature — Sliver {5}{R}, 3/3 (Time Spiral, #5596):
//
//	"All Sliver creatures have double strike."
//
// A layer-6 keyword grant (CR 613.1f) of double strike (CR 702.4a) to
// every Sliver creature at the table, an opponent's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44623693-51d6-49ad-8cd7-140505caf02f",
		Name:         "Fury Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "double strike"),
		},
	})
}
