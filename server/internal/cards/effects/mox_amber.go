package effects

// Mox Amber — Legendary Artifact {0}:
//
//	"{T}: Add one mana of any color among legendary creatures and
//	 planeswalkers you control."
//
// Rank 204. The third derived-mana shape, and the one that reads
// permanents' COLOURS rather than what they could produce: a legendary
// green creature offers {G} whether or not it taps for anything.
//
// A free artifact that does nothing until a legend lands, which is why
// it is played in commander decks and almost nowhere else — the
// commander itself is the enabler, and it is legendary by definition.
//
// "Any color", so colorless legends contribute nothing and a board of
// only colorless legends makes this produce no mana at all. Same
// empty-string outcome as Exotic Orchard with no opposing lands: the
// artifact taps, nothing arrives.
//
// Reads Effective() supertypes and types, so a creature that is
// legendary only by way of a continuous effect counts, and the colours
// come from EffectiveColors so a colour-changing effect is respected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "7a43bd27-fdd8-41f0-9bc4-92568f3408f1",
		Name:     "Mox Amber",
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedFromLegendaryPermanents(),
			Label:        "Add one mana of any color among legendary creatures and planeswalkers you control",
			// The colours come from the board, not from the command
			// zone — narrowing a derived set by commander identity
			// would be a second, unprinted filter.
			IgnoreCommanderIdentity: true,
		}},
	})
}
