package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "39d283e0-9d9b-4c04-aa6a-66a9883af9a2",
		Name:            "Toucan-Puffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Toucan-Puffin — target creature you control gets +2/+0 until end of turn",
				boostTargetUntilEOT("Toucan-Puffin — +2/+0", 2, 0)),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
