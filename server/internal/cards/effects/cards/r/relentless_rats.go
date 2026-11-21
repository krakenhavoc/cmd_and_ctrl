package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relentless Rats — Creature — Rat {1}{B}{B}, 2/2 (Special Guests):
//
//	"This creature gets +1/+1 for each other creature on the battlefield
//	 named Relentless Rats.
//	 A deck can have any number of cards named Relentless Rats."
//
// A layer 7c self-pump (CR 613.4c). Unlike Rat Colony it counts every
// player's creatures, and it counts by name as the layers leave it, so a
// copy of a Relentless Rats counts and a Rat with another name does not.
// The deck rule is CR 113.6n's, read off the oracle text by internal/deck.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "104ea189-14cd-420f-afdc-57b0f827ab8e",
		Name:         "Relentless Rats",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Gets +1/+1 for each other creature on the battlefield named Relentless Rats.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, o := range g.BattlefieldCardsForEffect() {
					if o.InstanceID != source.InstanceID && o.IsCreature() && game.HasName(o, "Relentless Rats") {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
