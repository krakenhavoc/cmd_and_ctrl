package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bestow.go — the card side of CR 702.103, bestow (ADR 0141, #2862).
// The rules side is game/bestow.go. Append-only, per the shared-
// vocabulary rule in docs/adding-cards.md.
//
// A bestow card is an enchantment creature with one more alternative
// cost and statics that name "enchanted creature":
//
//	AlternativeCosts: []game.AlternativeCost{Bestow("{3}{G}{G}")},
//	Static: []game.StaticAbility{PumpAttached(4, 2)},
//
// PumpAttached and GrantToAttached already apply to whatever the
// permanent is attached to, so a bestowed Aura pumps its host and an
// unattached creature pumps nothing. A static printed as "this
// creature and enchanted creature each get …" is SelfCreatureOrAttached.

// Bestow is "Bestow {cost} (If you cast this card for its bestow cost,
// it's an Aura spell with enchant creature. It becomes a creature
// again if it's not attached.)" — CR 702.103a.
//
// The price, the enchant creature clause and the bit that makes the
// spell an Aura, in one value, for the reason Evoke bundles its
// sacrifice: a card file that wrote the cost by hand would cast a
// creature spell for the bestow cost with a target it never attaches
// to. The clause is the one an ordinary Aura's Spec.Targets carries
// (EnchantCreature), so the announce gate, the CR 608.2b re-check, the
// view's per-offer legal set and the bot enumerator read it as they
// read cleave's.
func Bestow(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:      game.BestowKey,
		Label:    "Bestow " + cost,
		ManaCost: cost,
		Targets:  EnchantCreature(),
		Bestow:   true,
	}
}

// SelfCreatureOrAttached is the AppliesTo predicate for "this creature
// and enchanted creature each get …" (Nighthowler, Eidolon of Countless
// Battles): the permanent itself while it is a creature, and the
// creature it is attached to while it is a bestowed Aura. Read during
// the layer pass, after layer 4, so a bestowed Aura is not a creature
// here and the bonus lands on its host alone.
func SelfCreatureOrAttached(target *game.Card, g *game.Game, source *game.Card) bool {
	if selfOnly(target, g, source) {
		return target.IsCreature()
	}
	return AttachedToSource(target, g, source)
}

// PumpSelfCreatureOrAttachedPer is "This creature and enchanted creature
// each get +P/+T for each <thing>": a layer 7c modify over
// SelfCreatureOrAttached, `count` read off the source on every pass.
func PumpSelfCreatureOrAttachedPer(power, toughness int, count func(g *game.Game, source *game.Card) int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfCreatureOrAttached,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := count(g, source)
			c.Power += power * n
			c.Toughness += toughness * n
		},
	}
}
