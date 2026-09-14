package effects

// Grove of the Burnwillows — Land (EDHREC rank 3354):
//
//	"{T}: Add {C}.
//	 {T}: Add {R} or {G}. Each opponent gains 1 life."
//
// The Gruul painland whose "pain" is paid to the table. Two mana
// abilities in the painland shape: the painless {C} first, so the
// auto-tapper reaches for it, and the coloured pair with a rider —
// the opponents' life gain, through the ordinary life-change path so
// their lifegain payoffs see it, as printed. The pipe is not
// narrowed to the commander's identity: the card names two colours
// and says nothing about the command zone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d33c3fbb-8306-4c2d-b0dd-88f12639da94",
		Name:         "Grove of the Burnwillows",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:                    ManaAbilityCost{Tap: true},
				Produced:                "{R|G}",
				Label:                   "Add {R} or {G}. Each opponent gains 1 life.",
				Rider:                   b32EachOpponentGainsLifeRider(1),
				IgnoreCommanderIdentity: true,
			},
		},
	})
}
