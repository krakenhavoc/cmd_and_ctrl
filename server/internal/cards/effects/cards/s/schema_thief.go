package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Schema Thief — Creature — Vedalken Rogue Artificer {3}{U}, 3/3 (March of
// the Machine Commander):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, create a token
//	 that's a copy of target artifact that player controls."
//
// The damage trigger names the damaged player's artifacts in its target
// clause (ControlledByThatPlayer); the token is a copy of the artifact's
// copiable values (CR 707.2), made under your control.
//
// No simplification.
func init() {
	damage := effects.WheneverThisDealsCombatDamageToAPlayer(
		"Schema Thief — create a token that's a copy of target artifact that player controls",
		effects.ExtravagantReplicationCopy)
	damage.Targets = effects.ControlledByThatPlayer(effects.TargetPermanent("target artifact that player controls", effects.Artifact()))
	effects.Register(effects.Spec{
		OracleID:        "9495b2b4-21bf-488b-a64e-74592bb277d9",
		Name:            "Schema Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{damage},
	})
}
