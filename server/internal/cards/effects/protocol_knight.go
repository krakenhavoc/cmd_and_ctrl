package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Protocol Knight — Creature — Human Knight {3}{U}, 3/4 (March of the
// Machine):
//
//	"When this creature enters, tap target creature an opponent controls.
//	 Put a stun counter on that creature if you control another Knight."
//
// The Knight test is made as the ability resolves; it is a resolution
// condition, not an intervening "if".
func init() {
	Register(Spec{
		OracleID:     "ac935fad-2fa7-4afc-9db4-3cd30ec4aabd",
		Name:         "Protocol Knight",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Protocol Knight — tap target creature an opponent controls; stun counter if you control another Knight", s82Mom09TapAndMaybeStun),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
