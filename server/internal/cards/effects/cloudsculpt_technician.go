package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudsculpt Technician — Creature — Jellyfish Artificer {2}{U}, 1/4
// (Edge of Eternities, #4461):
//
//	"Flying
//	 As long as you control an artifact, this creature gets +1/+0."
//
// Flying is a printed keyword. The pump is a layer 7c static (CR 613.4c) on
// the Technician alone, on while its controller controls an artifact. The
// check reads post-layer types (BattlefieldCardsForEffect, as Doctor Doom's
// does), so a permanent made an artifact by another effect counts, and the
// pump ends the moment the last artifact leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c5183be-cddf-48fb-a4b5-dc5f6faf26de",
		Name:            "Cloudsculpt Technician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "As long as you control an artifact, this creature gets +1/+0.",
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && cloudsculptControlsAnArtifact(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
	})
}

// cloudsculptControlsAnArtifact is "you control an artifact".
func cloudsculptControlsAnArtifact(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsArtifact() {
			return true
		}
	}
	return false
}
