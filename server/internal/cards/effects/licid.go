package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// licid.go — the licids (#5781, ADR 0154 amendment 2026-11-25,
// game/licid.go). A licid's card declares three things:
//
//	ReturnedAura: EnchantCreature(),
//	Static:       append(LicidForm(0), <what the Aura does>...),
//	Activated:    []ActivatedAbility{LicidAbility("{B}", "{B}")},
//
// LicidAbility is the printed ability; LicidForm is what it turns the
// permanent into (the type set and the loss of that ability, read off
// Card.ReturnedAura); ReturnedAura is "with enchant creature", the clause
// CR 704.5m holds the Aura to. The Aura's own text ("Enchanted creature
// has haste", "At the beginning of the upkeep of enchanted creature's
// controller, …") is written with the ordinary Aura helpers: it refers to
// nothing while the licid is a creature, because a creature enchants
// nothing. Paying to end the effect is the engine's end_effect special
// action (CR 116.2c), derived from the permanent, so the card declares
// nothing for it.

// LicidAbility is "{cost}, {T}: This creature loses this ability and
// becomes an Aura enchantment with enchant creature. Attach it to target
// creature. You may pay {endCost} to end this effect." It must be the
// card's activated row `row` that LicidForm names (the first, on every
// licid).
//
// Nothing happens if the licid has left the battlefield or become a new
// object since it was activated (CR 400.7), and the ability is removed
// from the stack if its target is illegal (CR 608.2b), so it never
// becomes an Aura of nothing. "You" is the ability's controller
// (CR 109.5), who may end the effect whoever controls the licid later.
func LicidAbility(cost, endCost string) ActivatedAbility {
	return ActivatedAbility{
		Label: cost + ", {T}: This creature loses this ability and becomes an Aura enchantment with enchant creature. Attach it to target creature. You may pay " +
			endCost + " to end this effect.",
		Cost:    Plus(ManaCost(cost), TapCost()),
		Targets: TargetCreature("target creature"),
		Effect: func(g *game.Game, item *game.StackItem) error {
			if !SourceIsStillThisPermanent(g, item) {
				return nil
			}
			host, ok := B16FirstLegalTargetCard(NewContext(g, item))
			if !ok {
				return nil
			}
			_, err := g.BecomeLicidAuraForEffect(item.SourceCardID, item.Controller, endCost, host)
			return err
		},
	}
}

// LicidForm is the licid's Aura form, applied while Card.ReturnedAura is
// set: it is an Aura enchantment and nothing else (CR 205.1a: "becomes an
// Aura enchantment" sets its card types, so it is no creature and has no
// creature types, layer 4), and it has lost its activated row `row`
// (layer 6, CR 613.1f), the licid ability. Both end with the effect.
func LicidForm(row int) []game.StaticAbility {
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID && source.ReturnedAura
	}
	removal := game.OwnAbilityRemoval{Slot: game.AbilitySlotActivated, Row: row}
	return []game.StaticAbility{
		{
			Label:     "It's an Aura enchantment.",
			Layer:     game.Layer4Type,
			AppliesTo: self,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Types = []string{"Enchantment"}
				c.SetSubtypes([]string{"Aura"})
			},
		},
		{
			Label:     "It loses this ability.",
			Layer:     game.Layer6Ability,
			AppliesTo: self,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, r := range c.RemovedOwnAbilities {
					if r == removal {
						return
					}
				}
				c.RemovedOwnAbilities = append(c.RemovedOwnAbilities, removal)
			},
		},
	}
}
