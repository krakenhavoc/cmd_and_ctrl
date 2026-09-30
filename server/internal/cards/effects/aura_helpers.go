package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// aura_helpers.go — counting helpers for Auras whose bonus is "for
// each <thing>" (PumpAttachedPer's `count`). Append-only, per the
// shared-vocabulary rule in docs/adding-cards.md.
//
// All of them run inside a layer recompute where the caller may hold
// only the read lock, so they walk g.Battlefield.Cards directly.

// enchantmentsControlledBy counts "each enchantment you control" —
// Ethereal Armor. The Aura itself is one, and a permanent that is an
// enchantment only through an earlier layer still counts (Effective
// types, not printed).
func enchantmentsControlledBy(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool {
		for _, t := range c.Effective().Types {
			if t == "Enchantment" {
				return true
			}
		}
		return false
	})
}

// aurasYouControlOnCreatures counts "each Aura you control that's
// attached to a creature" — Sage's Reverie. "You control" is the
// SOURCE's controller; the Aura may be on anyone's creature, and an
// Aura on a land or a player does not count.
func aurasYouControlOnCreatures(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.Controller != source.Controller || !a.IsAura() {
			continue
		}
		if host := g.AttachedHostOf(a); host != nil && host.IsCreature() {
			n++
		}
	}
	return n
}

// aurasAndEquipmentOnTheHost counts "each Aura and Equipment attached
// to it" for an Aura's own host — Mantle of the Ancients. Everybody's
// attachments count, the source included, because the printed text
// names no controller.
func aurasAndEquipmentOnTheHost(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	host := g.AttachedHostOf(source)
	if host == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAttachedTo(host.InstanceID) && (a.IsAura() || a.HasSubtype("Equipment")) {
			n++
		}
	}
	return n
}

// untapAttachedHost is "untap enchanted / equipped creature" for an
// attachment's own ability: the host is found through the attachment
// as the ability resolves, and an attachment that has fallen off by
// then untaps nothing. Sting, the Glinting Dagger's combat trigger
// and Freed from the Real's second ability.
func untapAttachedHost(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return UntapTarget{Target: host.InstanceID}.Apply(NewContext(g, item))
}

// tapAttachedHost is untapAttachedHost's twin, for "tap enchanted
// creature" (Freed from the Real).
func tapAttachedHost(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return TapTarget{Target: host.InstanceID}.Apply(NewContext(g, item))
}
