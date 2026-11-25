package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrash of Raptors — Creature — Dinosaur {3}{R}, 3/3 (Ixalan):
//
//	"As long as you control another Dinosaur, this creature gets +2/+0 and
//	 has trample."
//
// Hungry Ridgewolf's shape: a layer 7c +2/+0 and a layer 6 trample grant on
// the source alone, both gated on some other Dinosaur its controller controls,
// read live through the layers.
//
// No simplification.
func init() {
	applies := effects.OldB221SelfWhileAnother([]string{"Dinosaur"}, nil)
	effects.Register(effects.Spec{
		OracleID:     "5f554df8-c70f-4212-aec7-597fe8a88553",
		Name:         "Thrash of Raptors",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control another Dinosaur, this creature gets +2/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: applies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
				},
			},
			effects.KeywordGrant(applies, "trample"),
		},
	})
}
