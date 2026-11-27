package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talon Sliver — Creature — Sliver {1}{W}, 1/1 (Tempest, S82 OLD, #5628):
//
//	"All Sliver creatures have first strike."
//
// Winged Sliver's grant with first strike, one layer-6 keyword over every Sliver
// creature at the table (CR 613.1f), this one and an opponent's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74910702-e385-48f7-a6b3-064ebed08be7",
		Name:         "Talon Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "first strike"),
		},
	})
}
