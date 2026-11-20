package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vulshok Heartstoker — Creature — Human Shaman {2}{R}, 2/2 (Scars of Mirrodin, #5577):
//
//	"When this creature enters, target creature gets +2/+0 until end of
//	 turn."
//
// Daybreak Charger's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d996788-7193-45cd-bc6d-0f470521d1ca",
		Name:         "Vulshok Heartstoker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Vulshok Heartstoker — target creature gets +2/+0 until end of turn",
				boostTargetUntilEOT("Vulshok Heartstoker — +2/+0", 2, 0)),
				TargetCreature("target creature")),
		},
	})
}
