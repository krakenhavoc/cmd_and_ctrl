package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rubblebelt Boar — Creature — Boar {3}{R}, 3/3 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, target creature gets +2/+0 until end of turn."
//
// Daybreak Charger's targeted enter trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b815cbbd-d79f-4180-8204-504a15df7152",
		Name:         "Rubblebelt Boar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Rubblebelt Boar — target creature gets +2/+0 until end of turn",
				boostTargetUntilEOT("Rubblebelt Boar — +2/+0", 2, 0)),
				TargetCreature("target creature")),
		},
	})
}
