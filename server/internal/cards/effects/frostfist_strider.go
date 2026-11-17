package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frostfist Strider — Creature — Elemental Giant {3}{U}{U}, 4/4 (Dominaria
// United, #5487):
//
//	"Ward {2}
//	 When this creature enters, tap target creature an opponent controls
//	 and put a stun counter on it. (If a permanent with a stun counter
//	 would become untapped, remove one from it instead.)"
//
// Ward is the shared trigger. The enters trigger is Cryotheory Adept's tap
// and stun on a target chosen as the trigger goes on the stack, limited to
// creatures an opponent controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e4a21a8-9332-4aeb-8519-949feb211294",
		Name:         "Frostfist Strider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Frostfist Strider — ward {2}"),
			Targeting(WhenThisEnters("Frostfist Strider — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
