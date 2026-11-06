package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rowdy Snowballers — Creature — Human Peasant Ally {2}{U}, 2/2 (Avatar:
// The Last Airbender, #4459):
//
//	"When this creature enters, tap target creature an opponent controls
//	 and put a stun counter on it. (If a permanent with a stun counter
//	 would become untapped, remove one from it instead.)"
//
// Deluge Virtuoso's enters trigger: the shared tap-and-stun body
// (tapTargetAndStunIt), the target re-checked as it resolves (CR
// 608.2b). The stun counter is CR 122.1d's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74224cce-c2ce-4456-8a97-3abbbb199385",
		Name:         "Rowdy Snowballers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Rowdy Snowballers — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
