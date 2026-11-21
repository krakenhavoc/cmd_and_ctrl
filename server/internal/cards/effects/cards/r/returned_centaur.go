package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Returned Centaur — Creature — Zombie Centaur {3}{B}, 2/4 (Theros, #5561):
//
//	"When this creature enters, target player mills four cards."
//
// Wall of Lost Thoughts' trigger: any player may be the target, its
// controller included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "223c97fb-dc96-41ba-be9d-bad9b5e4876f",
		Name:         "Returned Centaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Returned Centaur — target player mills four cards", effects.TargetPlayerMills(4)),
				effects.TargetPlayer("target player")),
		},
	})
}
