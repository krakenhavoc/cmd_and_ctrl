package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elturel Survivors — Creature — Tiefling Peasant {3}{R}, 0/4 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Trample, myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and attacking that
//	 player or a planeswalker they control. Exile the tokens at end of combat.)
//	 As long as this creature is attacking, it gets +X/+0, where X is the number of
//	 lands defending player controls."
//
// Trample is printed and myriad is the shared trigger (CR 702.116). The last line is
// a layer 7c static (CR 613.4c) that applies only while this creature is attacking
// (DependsOnAttackingStatus, so the layers are read again as the attack is declared
// and when combat ends) and counts the lands controlled by the defending player of
// its attack, a player whose planeswalker or battle is attacked included
// (CR 508.5): each myriad token copy attacks a different player and reads its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "373ccd23-5c98-4368-995a-e6503cd0dbd8",
		Name:            "Elturel Survivors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "As long as this creature is attacking, it gets +X/+0, where X is the number of lands defending player controls.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				c.Power += effects.GapB219LandsDefendingPlayerControls(g, target)
			},
			DependsOnAttackingStatus: true,
		}},
		Triggered: []game.TriggeredAbility{effects.GrantedMyriad("Elturel Survivors")},
	})
}
