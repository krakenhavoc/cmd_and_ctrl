package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runaway Trash-Bot — Artifact Creature — Construct {3}, 0/4 (Kamigawa:
// Neon Dynasty, #5493):
//
//	"Trample
//	 This creature gets +1/+0 for each artifact and/or enchantment card in
//	 your graveyard."
//
// A layer 7c modify read on every pass (Carrion Grub's shape): each card
// in its controller's graveyard that is an artifact, an enchantment or
// both counts once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93b0bd0e-414b-45bd-9d3d-7323140ef0c0",
		Name:            "Runaway Trash-Bot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			Label:     "This creature gets +1/+0 for each artifact and/or enchantment card in your graveyard.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil || p.Graveyard == nil {
					return
				}
				for _, gc := range p.Graveyard.Cards {
					if gc.IsArtifact() || gc.IsEnchantment() {
						c.Power++
					}
				}
			},
		}},
	})
}
