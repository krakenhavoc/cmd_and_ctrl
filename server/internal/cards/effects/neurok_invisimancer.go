package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neurok Invisimancer — Creature — Human Wizard {1}{U}{U}, 2/1 (Scars of
// Mirrodin, #5577):
//
//	"This creature can't be blocked.
//	 When this creature enters, target creature can't be blocked this
//	 turn."
//
// Dimir Infiltrator's static and Clammy Prowler's "target creature can't
// be blocked this turn" (targetCantBeBlockedThisTurn) on a required
// enters trigger. The target may be the Invisimancer itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7425964-0028-4156-9037-b303b00e7e11",
		Name:         "Neurok Invisimancer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Neurok Invisimancer — target creature can't be blocked this turn",
					targetCantBeBlockedThisTurn("Neurok Invisimancer")),
				TargetCreature("target creature")),
		},
	})
}
