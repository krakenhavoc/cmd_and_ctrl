package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haazda Marshal — Creature — Human Soldier {W}, 1/1 (Guilds of Ravnica,
// #5525):
//
//	"Whenever this creature and at least two other creatures attack, create a
//	 1/1 white Soldier creature token with lifelink."
//
// Battalion (Boros Strike-Captain's group test: this creature and at least two
// other creatures declared as attackers in the same declaration) with Dawn of
// Hope's 1/1 white Soldier token with lifelink. The token is created whether or
// not the Marshal survives to resolution (its controller is the trigger's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a1efa7e-21d5-47a1-bcd8-7d4616be0d58",
		Name:         "Haazda Marshal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Haazda Marshal — create a 1/1 white Soldier creature token with lifelink",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Soldier with lifelink"), N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
