package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantasmal Dreadmaw — Creature — Dinosaur Illusion {2}{U}{U}, 6/6
// (Modern Horizons 2, #5501):
//
//	"Trample
//	 When this creature becomes the target of a spell or ability,
//	 sacrifice it."
//
// Phantasmal Image's Illusion trigger, printed here rather than granted:
// it uses the stack, and the Dreadmaw is sacrificed only if it is still
// on the battlefield as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cbde4b9d-1549-4892-b8ca-38256a6ac1ba",
		Name:            "Phantasmal Dreadmaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self,
				"Phantasmal Dreadmaw — sacrifice it",
				effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
