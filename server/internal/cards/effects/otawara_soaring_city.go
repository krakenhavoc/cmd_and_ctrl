package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Otawara, Soaring City — Legendary Land:
//
//	"{T}: Add {U}.
//	 Channel — {3}{U}, Discard this card: Return target artifact,
//	 creature, enchantment, or planeswalker to its owner's hand. This
//	 ability costs {1} less to activate for each legendary creature
//	 you control."
//
// Boseiju's blue sibling, and the same shape: an Island that is also
// an instant-speed bounce you never have to draw into. Channel (CR
// 702.142) is an ACTIVATED ability that functions from the hand, and
// `ActivatedAbility.Zones` (#660) is the slot that says so — with the
// `DiscardSelf` cost component and the CR 108.4 "the card's owner is
// you" rule both riding the one activation path.
//
// The bounce is not restricted to an opponent's permanents: Otawara
// can save your own commander from an exile effect, which is a real
// and printed use of the card.
//
// The discard is a COST (CR 602.2b), paid at announce with the
// ability on the stack, so a countered channel still costs the land.
//
// "This ability costs {1} less to activate for each legendary
// creature you control" is the channel ability's OWN cost clause
// (ActivatedAbility.CostModifiers, #1296), so it prices from the hand
// where channel is activated — a board modifier would never be found
// there (CR 113.6). It used to be a declared simplification; see
// Takenuma, Abandoned Mire for why the #1184 board route did not
// reach it either.
func init() {
	Register(Spec{
		OracleID:     "e9b6a394-691c-425a-9307-76d8edc7375e",
		Name:         "Otawara, Soaring City",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:         "Channel — {3}{U}, Discard this card: Return target artifact, creature, enchantment, or planeswalker to its owner's hand",
			Cost:          game.AbilityCost{Mana: "{3}{U}", DiscardSelf: true},
			Zones:         []game.ZoneKind{game.ZoneHand},
			CostModifiers: []game.CostModifier{ChannelDiscountPerLegendaryCreature()},
			Targets: TargetPermanent("target artifact, creature, enchantment, or planeswalker",
				Or(Artifact(), Creature(), Enchantment(), Planeswalker())),
			Effect: otawaraChannel,
		}},
	})
}

// otawaraChannel returns the chosen permanent to its owner's hand. A
// target that has gone is not an error (CR 608.2b).
func otawaraChannel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return BounceToHand{Target: t.ID}.Apply(ctx)
	}
	return nil
}
