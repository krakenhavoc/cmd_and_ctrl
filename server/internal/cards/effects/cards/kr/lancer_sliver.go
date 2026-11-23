package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lancer Sliver — Creature — Sliver {2}{W}, 2/2 (Modern Horizons, #5519):
//
//	"Sliver creatures you control have first strike."
//
// Sentinel Sliver's shape with first strike: no "other", so the Lancer
// has it too, and "you control" keeps it off an opponent's Slivers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ca33b31-4707-40d0-b4ed-cbe625793122",
		Name:         "Lancer Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "first strike"),
		},
	})
}
