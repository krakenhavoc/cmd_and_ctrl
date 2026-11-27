package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyshroud Elite — Creature — Elf {G}, 1/1 (Exodus, S82 OLD b740, #5626):
//
//	"This creature gets +1/+2 as long as an opponent controls a nonbasic
//	 land."
//
// A layer-7c modification (Malicious Invader's shape) that applies to the Elite
// alone while any other player controls a land without the basic supertype,
// read through the layers as the characteristics are computed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd9f2c84-0589-4ffe-9bff-ea74be4f77a4",
		Name:         "Skyshroud Elite",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+2 as long as an opponent controls a nonbasic land.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && skyshroudEliteOpponentHasNonbasicLand(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness += 2
			},
		}},
	})
}

// skyshroudEliteOpponentHasNonbasicLand reports whether a player other than
// `you` controls a land that is not basic.
func skyshroudEliteOpponentHasNonbasicLand(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.IsLand() && !c.HasSupertype("basic") {
			return true
		}
	}
	return false
}
