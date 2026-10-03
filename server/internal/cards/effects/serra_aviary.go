package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra Aviary — World Enchantment {3}{W}:
//
//	"Creatures with flying get +1/+1."
//
// Every player's fliers. Layer 7c, after layer 6 has settled which
// creatures have flying, so a creature that lost flying to a Gravity
// Sphere gets nothing and one granted flying gets the bonus.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a06483b-e71c-4d11-86ea-3b48e2a9cdaf",
		Name:         "Serra Aviary",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, _ *game.Card) bool {
				return target.IsCreature() && HasKeyword("flying")(g, uuid.Nil, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}
