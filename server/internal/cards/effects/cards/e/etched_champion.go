package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etched Champion — Artifact Creature — Soldier {3}, 2/2:
//
//	"Metalcraft — This creature has protection from each color as long
//	 as you control three or more artifacts."
//
// A layer-6 static on itself that adds one protection quality per colour
// (CR 702.16 — "protection from each color" is five separate abilities,
// Akroma's Will's shape) while the controller has three or more
// artifacts. The count is read as the layer pass reaches layer 6, so it
// switches on with the third artifact and off when one leaves; the
// Champion counts itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cd7102d-6413-4f46-9434-9130aefa2f88",
		Name:         "Etched Champion",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Metalcraft — This creature has protection from each color as long as you control three or more artifacts.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() }) >= 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, col := range game.AllColors {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, game.ProtectionFromColor(col))
				}
			},
		}},
	})
}
