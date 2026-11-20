package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protocol Knight — Creature — Human Knight {3}{U}, 3/4 (March of the
// Machine):
//
//	"When this creature enters, tap target creature an opponent controls.
//	 Put a stun counter on that creature if you control another Knight."
//
// The Knight test is made as the ability resolves; it is a resolution
// condition, not an intervening "if".
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac935fad-2fa7-4afc-9db4-3cd30ec4aabd",
		Name:         "Protocol Knight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Protocol Knight — tap target creature an opponent controls; stun counter if you control another Knight", effects.S82Mom09TapAndMaybeStun),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
