package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vile Manifestation — Creature — Horror {1}{B}, 0/4 (Hour of Devastation,
// #5534):
//
//	"This creature gets +1/+0 for each card with cycling in your graveyard.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Winter, Tormented Loner's graveyard-counting layer 7c bonus, read live
// through the layers, over the cards in your graveyard that have a cycling
// ability (CR 702.29a; typecycling counts too, CR 702.29e). The count is the
// graveyard's alone, so cycling Vile Manifestation itself adds a card with
// cycling and the creature the cycle came from (if on the battlefield)
// grows. Cycling is the shared activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d747c43-a09b-4209-b042-2b89156612d8",
		Name:         "Vile Manifestation",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += effects.S82OldB235CyclingCardsInGraveyard(g, source.Controller)
			},
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
