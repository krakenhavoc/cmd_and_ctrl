package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvage Slasher — Artifact Creature — Human Rogue {1}{B}, 1/1 (Conflux,
// #5587):
//
//	"This creature gets +1/+0 for each artifact card in your graveyard."
//
// A layer-7c modification (CR 613.4c) counted continuously from its
// controller's graveyard, Fiend Artisan's shape over artifact cards and
// power only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "196d6ef6-80d9-4bfd-a5ed-d9e8b00faad0",
		Name:         "Salvage Slasher",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+0 for each artifact card in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					for _, gc := range p.Graveyard.Cards {
						if gc.IsArtifact() {
							c.Power++
						}
					}
				}
			},
		}},
	})
}
