package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timberpack Wolf — Creature — Wolf {1}{G}, 2/2 (Magic Origins, S82 older sets):
//
//	"This creature gets +1/+1 for each other creature you control named
//	 Timberpack Wolf."
//
// Relentless Rats' name count narrowed to the controller's creatures: a layer 7c
// self-pump (CR 613.4c) counting by name as the layers leave a creature, so a
// copy of a Timberpack Wolf counts and a creature with another name does not.
// A deck may hold any number of them (CR 113.6n, read off the oracle text).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9375dca-2b32-4185-b88a-4bd5dde9a5d6",
		Name:         "Timberpack Wolf",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 for each other creature you control named Timberpack Wolf.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.CountControlled(g, source.Controller, func(o game.Card) bool {
					return o.InstanceID != source.InstanceID && o.IsCreature() && game.HasName(o, "Timberpack Wolf")
				})
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
