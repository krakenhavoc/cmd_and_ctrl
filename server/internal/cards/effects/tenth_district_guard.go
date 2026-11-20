package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tenth District Guard — Creature — Human Soldier {1}{W}, 2/2 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, target creature gets +0/+1 until end of turn."
//
// Daybreak Charger's targeted enter trigger with a +0/+1 pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83ce0859-cb88-45dd-8485-3317792235de",
		Name:         "Tenth District Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Tenth District Guard — target creature gets +0/+1 until end of turn",
				boostTargetUntilEOT("Tenth District Guard — +0/+1", 0, 1)),
				TargetCreature("target creature")),
		},
	})
}
