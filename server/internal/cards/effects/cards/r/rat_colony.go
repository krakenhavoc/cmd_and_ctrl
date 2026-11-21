package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rat Colony — Creature — Rat {1}{B}, 2/1 (Special Guests):
//
//	"This creature gets +1/+0 for each other Rat you control.
//	 A deck can have any number of cards named Rat Colony."
//
// A layer 7c self-pump (CR 613.4c) counting the other Rats its controller
// controls through the layered view, so a changeling counts and a Rat an
// opponent controls does not. The deck rule is CR 113.6n's, read off the
// oracle text by internal/deck for every card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec77d23b-0165-450d-9aae-73b755163753",
		Name:         "Rat Colony",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Gets +1/+0 for each other Rat you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += effects.HobA9OtherSubtypeYouControl(g, source, "Rat")
			},
		}},
	})
}
