package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kodama of the West Tree — Legendary Creature — Spirit {2}{G}, 3/3
// (EDHREC rank 841):
//
//	"Reach
//	 Modified creatures you control have trample. (Equipment, Auras
//	 you control, and counters are modifications.)
//	 Whenever a modified creature you control deals combat damage to
//	 a player, search your library for a basic land card, put it onto
//	 the battlefield tapped, then shuffle."
//
// The +1/+1 counters deck's ramp engine. Trample is a Layer 6 grant
// over the creatures you control that are modified, recomputed every
// pass so a creature that gains its first counter mid-combat gets
// trample for the damage step; the ramp is a combat-damage trigger
// (Bident of Thassa's shape) gated on the dealer being modified, one
// basic per connecting creature, as printed.
//
// "Modified" is the full CR 700.9 word since #379 gave the engine an
// attachment relation: a counter of any kind, an Equipment attached,
// or an Aura the creature's controller controls attached
// (b07IsModified). The layer listener bumps on attach and unattach,
// so an Equipment moving off a creature drops its trample on the
// same pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d69b1e68-8d8e-460b-9eb4-6a68be886197",
		Name:            "Kodama of the West Tree",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Modified creatures you control have trample.",
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && effects.B07IsModified(g, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !effects.EotHasAbility(c.Abilities, "trample") {
					c.Abilities = append(c.Abilities, "trample")
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ModifiedCreatureDealtCombatDamageToPlayer, "Kodama of the West Tree — search for a basic land", func(g *game.Game, item *game.StackItem) error {
				return effects.B07SearchBasicOntoBattlefield(g, item, item.Controller, true, false,
					"Kodama of the West Tree — a basic land, tapped")
			}),
		},
	})
}
