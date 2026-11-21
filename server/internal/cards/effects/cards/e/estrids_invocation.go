package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const estridsInvocationBlink = "estrids-invocation/upkeep-blink"

// Estrid's Invocation — Enchantment {2}{U} (Modern Horizons 3, a reversible
// card with the same text on both faces):
//
//	"You may have this enchantment enter as a copy of an enchantment you
//	 control, except it has "At the beginning of your upkeep, you may exile
//	 this enchantment. If you do, return it to the battlefield under its
//	 owner's control.""
//
// Copy Enchantment's entry replacement over the enchantments YOU control
// (any enchantment, an Aura choosing what it enchants
// as it enters), with Colossal Clone Master's exception shape: the upkeep trigger
// is an ADR 0093 bundle granted to the copy as part of its copiable values (CR
// 707.9a). The trigger is Ghostly Castigator's optional trigger; "return it" is
// the shared flicker (Flicker), which returns the card, a new object, under its
// OWNER's control, so it enters again and chooses again what to copy. With no
// enchantment to copy, or on a decline, it enters as the plain Invocation, which
// does nothing.
//
// No simplification.
func init() {
	const label = "Estrid's Invocation — you may exile this enchantment; if you do, return it to the battlefield under its owner's control"
	blink := effects.Optional(effects.AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		if !effects.OnBattlefield(g, item.SourceCardID) {
			return nil
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok {
			return nil
		}
		return effects.Flicker{Target: item.SourceCardID, Controller: c.Owner}.Apply(effects.NewContext(g, item))
	}), label+"?")
	effects.Register(effects.Spec{
		OracleID:     "7f5f10ee-6428-4053-8824-3bc70420ec2c",
		Name:         "Estrid's Invocation",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       estridsInvocationBlink,
			Triggered: []game.TriggeredAbility{blink},
			Text:      "At the beginning of your upkeep, you may exile this enchantment. If you do, return it to the battlefield under its owner's control.",
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Estrid's Invocation",
				func(g *game.Game, controller, self uuid.UUID) []uuid.UUID {
					return effects.CopyCandidates(g, self, func(c game.Card) bool { return c.IsEnchantment() && c.Controller == controller })
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(estridsInvocationBlink)
				},
			),
		},
	})
}
