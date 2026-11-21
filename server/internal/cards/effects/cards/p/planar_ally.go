package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planar Ally — Creature — Angel {3}{W}{W}, 3/3:
//
//	"Flying
//	 Whenever this creature attacks, venture into the dungeon."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9ae30bd5-b638-4647-b59d-ffe08e77c182",
		Name:            "Planar Ally",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(effects.VentureLabel("Planar Ally"), effects.Venture),
		},
	})
}
