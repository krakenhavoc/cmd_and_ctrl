package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graceblade Artisan — Creature — Human Monk {2}{W}, 2/3 (Dragons of Tarkir, #5551):
//
//	"This creature gets +2/+2 for each Aura attached to it."
//
// Kor Spiritdancer's static, a layer 7c count of the Auras attached to it.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fa4e241-6b2f-4372-8bb6-a876e5ae7c50",
		Name:         "Graceblade Artisan",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Gets +2/+2 for each Aura attached to it.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				n := effects.B18AurasAttachedTo(g, target.InstanceID)
				c.Power += 2 * n
				c.Toughness += 2 * n
			},
		}},
	})
}
