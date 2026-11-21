package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toucan-Puffin — Creature — Bird {2}{W}, 2/2 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Flying
//	 When this creature enters, target creature you control gets +2/+0
//	 until end of turn."
//
// Daybreak Charger's enters trigger, narrowed to a creature you control
// (the Toucan-Puffin itself is a legal pick).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "39d283e0-9d9b-4c04-aa6a-66a9883af9a2",
		Name:            "Toucan-Puffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Toucan-Puffin — target creature you control gets +2/+0 until end of turn",
				effects.BoostTargetUntilEOT("Toucan-Puffin — +2/+0", 2, 0)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
