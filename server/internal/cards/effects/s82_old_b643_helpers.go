package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b643_helpers.go — shared bodies for Planeshift's enters-trigger cards,
// two-colour discounts and the land-type cards, batch 643 (#5615).

// OldB643CreatureOfEitherColor is the noun of "a blue or black creature you
// control": a creature that is either colour, or both. The colours are words of
// the asking effect's text, so ReturnOneYouControl reads them through it (CR 612.2).
func OldB643CreatureOfEitherColor(a, b string) func(game.Card) bool {
	return func(c game.Card) bool { return c.IsCreature() && (c.HasColorWord(a) || c.HasColorWord(b)) }
}

// OldB643EnchantmentOfEitherColor is "a red or green enchantment you control".
func OldB643EnchantmentOfEitherColor(a, b string) func(game.Card) bool {
	return func(c game.Card) bool { return c.IsEnchantment() && (c.HasColorWord(a) || c.HasColorWord(b)) }
}

// OldB643EachPlayerDiscards is "each player discards a card": every player
// chooses their own discard (CR 701.9a).
func OldB643EachPlayerDiscards(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
			Source: item.SourceCardID, N: 1, Question: name + " — discard a card",
		})
		return nil
	}
}

// OldB643TuckFirstTarget puts the first legal target card on top of its owner's
// library (CR 608.2b: a target that is no longer legal is not moved).
func OldB643TuckFirstTarget(g *game.Game, item *game.StackItem) error {
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind == game.TargetCard {
			return g.TuckToLibraryForEffect(t.ID, false)
		}
	}
	return nil
}

// OldB643SpellsOfEitherColorCostLess is "<A> spells and <B> spells you cast cost
// {1} less to cast" (the Planeshift familiars): a spell of either colour, or
// both. The colours are words of the source's text (CR 612.2).
func OldB643SpellsOfEitherColorCostLess(label, a, b string) game.CostModifier {
	return CostsLess(1, label, YourSpell(), func(q game.CostQuery) bool {
		return q.Card.HasColor(q.Source.TextColor(a)) || q.Card.HasColor(q.Source.TextColor(b))
	})
}

// OldB643YourLandsAre22FirstStrikeCreatures is Natural Emergence's static: the
// lands its controller controls are 2/2 creatures with first strike and still
// lands. Layer 4 adds the creature type and keeps every other, layer 7b sets the
// base 2/2 so counters and pumps still apply, layer 6 grants first strike.
func OldB643YourLandsAre22FirstStrikeCreatures(label string) []game.StaticAbility {
	yours := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsLand() && target.Controller == source.Controller
	}
	return []game.StaticAbility{
		{
			Label:     label,
			Layer:     game.Layer4Type,
			AppliesTo: yours,
			Apply:     OldB643AddCreatureType,
		},
		{
			Label:     label,
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: yours,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power, c.Toughness = 2, 2
			},
		},
		{
			Label:     label,
			Layer:     game.Layer6Ability,
			AppliesTo: yours,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, source.TextKeyword("first strike"))
			},
		},
	}
}

// OldB643AddCreatureType is a layer 4 Apply that adds the creature type and keeps
// every other type and subtype (Living Plane, Natural Emergence).
func OldB643AddCreatureType(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, t := range c.Types {
		if EqualFoldASCIIEffects(t, "Creature") {
			return
		}
	}
	c.Types = append(c.Types, "Creature")
}
