package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grappling Kraken — Creature — Kraken {4}{U}{U}, 5/6 (Foundations, #4465):
//
//	"Landfall — Whenever a land you control enters, tap target creature an
//	 opponent controls and put a stun counter on it. (If a permanent with a
//	 stun counter would become untapped, remove one from it instead.)"
//
// Landfall with Deluge Virtuoso's tap-and-stun body (tapTargetAndStunIt):
// one trigger per land, each with its own target, chosen as it goes on the
// stack. A target gone or no longer an opponent's by resolution gets
// nothing (CR 608.2b). The stun counter is CR 122.1d's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ba99ade-df8d-4654-882a-a5f4256ada3d",
		Name:         "Grappling Kraken",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Landfall("Grappling Kraken — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
