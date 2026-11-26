package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Resurgence — Creature — Wall {2}{W}, 0/6 (Oath of the Gatewatch,
// #5545):
//
//	"Defender
//	 When this creature enters, you may put three +1/+1 counters on target land
//	 you control. If you do, that land becomes a 0/0 Elemental creature with
//	 haste that's still a land."
//
// Awaken's clause as an enters trigger (CR 702.113a), Cyclone Sire's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8771e241-e69b-4e21-9108-223a9488e9c7",
		Name:            "Wall of Resurgence",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Wall of Resurgence — you may put three +1/+1 counters on target land you control; it becomes a 0/0 Elemental creature with haste", effects.OldB299ThreeCountersAndAnimateTheLand),
				effects.AwakenLandTargets()),
				"Wall of Resurgence — put three +1/+1 counters on target land you control?"),
		},
	})
}
