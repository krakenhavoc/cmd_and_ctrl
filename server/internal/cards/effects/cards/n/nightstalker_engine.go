package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nightstalker Engine — Creature — Nightstalker {4}{B}, */3 (Portal Second
// Age, S82 older sets batch 733, #5625):
//
//	"Nightstalker Engine's power is equal to the number of creature cards in
//	 your graveyard."
//
// A characteristic-defining ability (CR 604.3, layer 7a), self-only, counting
// the creature cards in its controller's graveyard live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "128b6751-4b47-4d7d-895e-c5d85776f6c4",
		Name:         "Nightstalker Engine",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of creature cards in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					for _, gc := range p.Graveyard.Cards {
						if gc.IsCreature() {
							n++
						}
					}
				}
				c.Power = n
			},
		}},
	})
}
