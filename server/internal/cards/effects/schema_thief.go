package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	damage := WheneverThisDealsCombatDamageToAPlayer(
		"Schema Thief — create a token that's a copy of target artifact that player controls",
		extravagantReplicationCopy)
	damage.Targets = ControlledByThatPlayer(TargetPermanent("target artifact that player controls", Artifact()))
	Register(Spec{
		OracleID:        "9495b2b4-21bf-488b-a64e-74592bb277d9",
		Name:            "Schema Thief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{damage},
	})
}
