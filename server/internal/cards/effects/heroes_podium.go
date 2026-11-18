package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heroes' Podium — Legendary Artifact {5} (Commander Legends reprint):
//
//	"Each legendary creature you control gets +1/+1 for each other
//	 legendary creature you control.
//	 {X}, {T}: Look at the top X cards of your library. You may reveal a
//	 legendary creature card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// The first line is a layer 7c static whose bonus is the number of OTHER
// legendary creatures the controller controls, read post-layer
// (s82RprB17OtherLegendaryCreatures). The second is Horn of the Mark's
// sentence (LookAtTopThenMayTakeToHand) with X paid; X = 0 looks at nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85292bfb-32d7-449e-9ebb-5be828e8ecf6",
		Name:         "Heroes' Podium",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Each legendary creature you control gets +1/+1 for each other legendary creature you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && isLegendary(target)
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
				n := s82RprB17OtherLegendaryCreatures(g, source.Controller, target.InstanceID)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: Look at the top X cards of your library. You may reveal a legendary creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := NewContext(g, item).X()
				if x <= 0 {
					return nil
				}
				return LookAtTopThenMayTakeToHand(x, And(Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLegendary() }), 1,
					"Heroes' Podium — reveal a legendary creature card and put it into your hand")(g, item)
			},
		}},
	})
}
