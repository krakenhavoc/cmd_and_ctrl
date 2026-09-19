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
// DECLARED SIMPLIFICATION, weaker than printed: "This ability costs
// {1} less to activate for each legendary creature you control" is
// not applied — cost modification reaches spells only, and an
// activated ability never passes through the CR 601.2f-style pricing
// pass (the open "Cost modification for activated abilities" seam).
// The channel always costs the printed {3}{U}, which is the
// acceptable direction (#259). Boseiju, Who Endures records the same
// gap.
func init() {
	Register(Spec{
		OracleID:     "e9b6a394-691c-425a-9307-76d8edc7375e",
		Name:         "Otawara, Soaring City",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The channel ability always costs {3}{U}; it doesn't get cheaper for each legendary creature you control.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label: "Channel — {3}{U}, Discard this card: Return target artifact, creature, enchantment, or planeswalker to its owner's hand",
			Cost:  game.AbilityCost{Mana: "{3}{U}", DiscardSelf: true},
			Zones: []game.ZoneKind{game.ZoneHand},
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
