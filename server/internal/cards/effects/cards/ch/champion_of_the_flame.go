package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Champion of the Flame — Creature — Human Warrior {1}{R}, 1/1 (Commander
// Masters):
//
//	"Trample
//	 This creature gets +2/+2 for each Aura and Equipment attached to it."
//
// Heavenly Blademaster's count (Auras and Equipment attached, whoever
// controls them), applied to the Champion itself in layer 7c (CR 613.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d9dbe40f-38e9-460a-b84f-4aa22650017e",
		Name:            "Champion of the Flame",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +2/+2 for each Aura and Equipment attached to it.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.GapB105AurasAndEquipmentAttached(g, source)
				c.Power += 2 * n
				c.Toughness += 2 * n
			},
		}},
	})
}
