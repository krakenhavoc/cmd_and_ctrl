package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Leotau — Creature — Cat {2}{G}{G}, 5/4 (Conflux, #5587):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {G}."
//
// Whipstitched Zombie's upkeep at {G}, through the shared
// UpkeepSacrificeUnlessYouPay: the controller is asked each upkeep, and
// declining (or being unable to pay) sacrifices it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5277fe0-b707-41bb-9ff9-62f8be249dd2",
		Name:         "Wild Leotau",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Wild Leotau", "{G}"),
		},
	})
}
