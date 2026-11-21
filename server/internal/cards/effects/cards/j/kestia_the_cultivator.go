package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kestia, the Cultivator — Legendary Enchantment Creature — Nymph
// {1}{G}{W}{U}, 4/4 (Commander 2018, #5526):
//
//	"Bestow {3}{G}{W}{U} (If you cast this card for its bestow cost, it's
//	 an Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Enchanted creature gets +4/+4.
//	 Whenever an enchanted creature or enchantment creature you control
//	 attacks, draw a card."
//
// Bestow and the pump are Boon Satyr's (ADR 0141). The draw is one trigger
// per creature you declared as an attacker that is an enchantment creature
// or has an Aura attached (b18AurasAttachedTo), read as it attacks: Kestia
// herself counts as an enchantment creature, and the creature she enchants
// as an Aura counts as an enchanted one. It works whether Kestia is a
// creature or attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21620bfc-8ef0-4e43-b675-d0bbad86e83d",
		Name:         "Kestia, the Cultivator",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{3}{G}{W}{U}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 4),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && (c.IsEnchantment() || effects.B18AurasAttachedTo(g, c.InstanceID) > 0)
			}, "Kestia, the Cultivator — an enchanted creature or enchantment creature you control attacked: draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
