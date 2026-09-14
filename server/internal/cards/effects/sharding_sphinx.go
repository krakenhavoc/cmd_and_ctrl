package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "9ebc0144-fc45-4de7-b3d8-ef8cf2e211ae",
		Name:            "Sharding Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b36ArtifactCreatureYouControlDealtCombatDamageToPlayer(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Sharding Sphinx — create a 1/1 blue Thopter with flying?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Sharding Sphinx — create a 1/1 blue Thopter with flying",
					b34CreateTokens(b36BlueThopterToken, 1))
			},
		}},
	})
}
