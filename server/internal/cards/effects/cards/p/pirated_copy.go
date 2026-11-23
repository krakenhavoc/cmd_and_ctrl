package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const piratedCopyDrawGrant = "pirated-copy/draw-on-same-name-damage"

// Pirated Copy — Creature — Shapeshifter Pirate {4}{U}, 0/0 (Jumpstart 2022, #5948):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it's a Pirate in addition to its other types and
//	 it has 'Whenever this creature or another creature with the same name
//	 deals combat damage to a player, you draw a card.'"
//
// Clone with a Pirate type and a granted triggered ability, both part of the
// copiable values (CR 707.9a, CR 707.9b), as on Phantasmal Image. The
// trigger fires for combat damage to a player from this creature or from any
// creature, whoever controls it, whose name matches this one's as the damage
// is dealt (names read through the layers). Its controller draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09ea5508-78fd-4bbc-8395-38ec13202d9a",
		Name:         "Pirated Copy",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  piratedCopyDrawGrant,
			Text: "Whenever this creature or another creature with the same name deals combat damage to a player, you draw a card.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.Source)
					return ok && c.IsCreature() && (c.InstanceID == source.InstanceID || c.Effective().Name == source.Effective().Name)
				}, "Pirated Copy — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Pirated Copy", effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Pirate")
					v.GrantAbility(piratedCopyDrawGrant)
				}),
		},
	})
}
