package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sharding Sphinx — Artifact Creature — Sphinx {4}{U}{U}, 4/4 (EDHREC
// rank 3817):
//
//	"Flying
//	 Whenever an artifact creature you control deals combat damage to
//	 a player, you may create a 1/1 blue Thopter artifact creature
//	 token with flying."
//
// The Thopter snowball: every artifact creature that connects makes
// another artifact creature. Flying rides PrintedKeywords. One
// optional trigger per artifact creature the controller controls
// that deals combat damage to a player — the Sphinx itself, its
// Thopters, an animated Treasure — read post-layer, so a creature
// something else made an artifact counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9ebc0144-fc45-4de7-b3d8-ef8cf2e211ae",
		Name:            "Sharding Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36ArtifactCreatureYouControlDealtCombatDamageToPlayer(ev, source, g)
			}, "Sharding Sphinx — create a 1/1 blue Thopter with flying", effects.B34CreateTokens(effects.B36BlueThopterToken, 1)), "Sharding Sphinx — create a 1/1 blue Thopter with flying?"),
		},
	})
}
