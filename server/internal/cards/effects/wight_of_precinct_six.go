package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wight of Precinct Six — Creature — Zombie {1}{B}, 1/1 (Jumpstart):
//
//	"This creature gets +1/+1 for each creature card in your opponents'
//	 graveyards."
//
// Wight of the Reliquary's layer-7c self boost, counted across every
// opponent's graveyard instead of yours. The count is read live, so it
// climbs as creatures die on the other side of the table and falls when a
// graveyard is exiled; your own graveyard never counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6397c046-4c59-4f0b-9b44-2a804eb95edf",
		Name:         "Wight of Precinct Six",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Gets +1/+1 for each creature card in your opponents' graveyards.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := jmpB3CreatureCardsInOpponentsGraveyards(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
