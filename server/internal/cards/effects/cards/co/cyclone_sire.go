package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cyclone Sire — Creature — Elemental {4}{U}, 3/4 (Oath of the Gatewatch,
// #5545):
//
//	"Flying
//	 When this creature dies, you may put three +1/+1 counters on target land
//	 you control. If you do, that land becomes a 0/0 Elemental creature with
//	 haste that's still a land."
//
// Awaken's clause as a trigger (CR 702.113a): the optional target land gets
// three counters and becomes a 0/0 Elemental creature with haste, with no
// duration (CR 611.2a). The land is judged as the trigger resolves; one that
// has left or changed hands does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b1c5278-76a9-4c96-b13a-0737aefb79d9",
		Name:            "Cyclone Sire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisDies("Cyclone Sire — you may put three +1/+1 counters on target land you control; it becomes a 0/0 Elemental creature with haste", effects.OldB299ThreeCountersAndAnimateTheLand),
				effects.AwakenLandTargets()),
				"Cyclone Sire — put three +1/+1 counters on target land you control?"),
		},
	})
}
