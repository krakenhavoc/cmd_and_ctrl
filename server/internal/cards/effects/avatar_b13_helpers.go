package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b13_helpers.go — shared helpers for Avatar: The Last Airbender,
// batch 13 (#4459).

// selfIsArtifactCreatureWhile is "As long as <cond>, this Vehicle is an
// artifact creature" (Phoenix Fleet Airship, War Balloon): a layer-4
// static on the source alone that adds the artifact and creature types
// while cond holds, read live on every layer pass. The Vehicle keeps its
// printed power and toughness.
func selfIsArtifactCreatureWhile(label string, cond func(g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && cond(g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, t := range []string{"Artifact", "Creature"} {
				if !slices.Contains(c.Types, t) {
					c.Types = append(c.Types, t)
				}
			}
		},
	}
}
