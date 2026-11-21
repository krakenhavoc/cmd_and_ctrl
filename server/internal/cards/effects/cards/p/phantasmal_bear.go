package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantasmal Bear — Creature — Bear Illusion {U}, 2/2 (Magic 2012, #5573):
//
//	"When this creature becomes the target of a spell or ability,
//	 sacrifice it."
//
// Phantasmal Shieldback's Illusion clause: a trigger on the Bear being
// targeted by any spell or ability, whose resolution sacrifices it if it
// is still the same permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "780da643-8d14-4f1a-a298-5b7454e9c4df",
		Name:         "Phantasmal Bear",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self,
				"Phantasmal Bear — sacrifice it",
				effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
