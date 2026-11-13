package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilded Scuttler — Artifact Creature — Crab {2}{U}, 1/3 (Foundations
// Jumpstart, #5464):
//
//	"This creature can't be blocked.
//	 When this creature enters, tap target creature an opponent controls
//	 and put a stun counter on it. (If a permanent with a stun counter
//	 would become untapped, remove one from it instead.)"
//
// Slither Blade's evasion and Mechanozoa's enters trigger, narrowed to a
// creature. A target gone by resolution gets nothing (CR 608.2b); with
// no creature an opponent controls the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "093f6123-6b1d-456f-a526-8740bc969b4e",
		Name:         "Gilded Scuttler",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Gilded Scuttler — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
