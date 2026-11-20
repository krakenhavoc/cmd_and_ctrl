package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana's Elite — Creature — Zombie {2}{B}, 1/1 (Eldritch Moon, #5541):
//
//	"This creature gets +1/+1 for each creature card in your graveyard."
//
// Wight of the Reliquary's characteristic-modifying static (layer 7c),
// counted on every recompute by b11CreatureCardsInGraveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6492082-d0ef-4f7c-a4b1-f8ecf6eb0890",
		Name:         "Liliana's Elite",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each creature card in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b11CreatureCardsInGraveyard(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
