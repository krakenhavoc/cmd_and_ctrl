package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sima Yi, Wei Field Marshal — Legendary Creature — Human Soldier {5}{B}, */4
// (Portal Three Kingdoms, #5622):
//
//	"Sima Yi's power is equal to the number of Swamps you control."
//
// The count is Dakmor Sorceress': it reads the effective type line, so a land
// that is a Swamp by any effect counts (CR 604.3, 613.4a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eefa9a94-9915-4dcd-b5dc-7f8e9aa621a7",
		Name:         "Sima Yi, Wei Field Marshal",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the number of Swamps you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.SwampsControlledBy(g, source.Controller, "Swamp")
			},
		}},
	})
}
