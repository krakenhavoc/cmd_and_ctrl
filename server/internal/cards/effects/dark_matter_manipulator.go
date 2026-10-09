package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Matter Manipulator — Creature — Human Warlock {B}, 1/2:
//
//	"When this creature enters, mill three cards.
//	 This creature gets +2/+0 for every seven cards in your
//	 graveyard."
//
// The static is a layer 7c modifier recomputed from the controller's
// graveyard on every pass, so milling past seven (or an exile that
// drops below it) moves the power at once. Cards in the graveyard are
// counted whatever their type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a3862f0-8dfc-4ac9-8eb9-faefcad685d4",
		Name:         "Dark Matter Manipulator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dark Matter Manipulator — mill three cards", Do(MillCards{N: 3})),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil || p.Graveyard == nil {
					return
				}
				c.Power += 2 * (len(p.Graveyard.Cards) / 7)
			},
		}},
	})
}
