package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indomitable Archangel — Creature — Angel {2}{W}{W}, 4/4:
//
//	"Flying
//	 Metalcraft — Artifacts you control have shroud as long as you
//	 control three or more artifacts."
//
// A layer-6 keyword grant to every artifact the controller has (the
// Archangel itself is not one), Privileged Position's shape gated on the
// artifact count as the layer pass reaches layer 6. The grant is not
// locked in (CR 611.3a): an artifact arriving later has shroud at once,
// and all of them lose it when the count drops below three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f9269eb-6a6c-4b32-8b3d-73724c12ca6b",
		Name:            "Indomitable Archangel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsArtifact() &&
					effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() }) >= 3
			}, "shroud"),
		},
	})
}
