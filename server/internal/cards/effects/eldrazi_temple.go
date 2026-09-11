package effects

// Eldrazi Temple — Land:
//
//	"{T}: Add {C}."
//	"{T}: Add {C}{C}. Spend this mana only to cast colorless Eldrazi
//	 spells or activate abilities of colorless Eldrazi."
//
// Rank 1720. The restricted-mana shape at its most demanding: two
// conditions on the object (colorless AND Eldrazi) and no condition on
// the KIND of payment — the mana funds a cast or an activation alike,
// which is unusual and is why the tag list here carries no
// ManaRestrictCast.
//
// The restriction is what makes the card printable at all. Unrestricted,
// "{T}: Add {C}{C}" on a land with no drawback is Ancient Tomb without
// the two damage — one of the strongest lands ever printed rather than
// a narrow tribal accelerant. Stamping the tags without enforcing them
// would ship exactly that, which is the #259 rule; the enforcement is
// in game/mana_restriction.go and it is why this card could not exist
// before #352.
//
// How the two tags resolve at spend time:
//
//   - "colorless" tests the object being paid for, so an Eldrazi with
//     a coloured mana cost (the Eldrazi Titans' devoid-less cousins,
//     and every coloured Eldrazi printed since Zendikar Rising) does
//     NOT qualify. That is the printed wording.
//   - "subtype:Eldrazi" reads the effective type line, so a creature
//     that is an Eldrazi only by way of a continuous effect counts.
//
// Both tags must hold (AND). A cast of a colorless non-Eldrazi
// artifact cannot use this mana, and neither can a colored Eldrazi.
//
// The first ability is ordinary colourless mana and auto-taps
// normally; the restricted ability does not (autoTapAbilityFor skips
// restricted abilities), so a player who wants the second {C} clicks
// for it. Weaker than printed, declared, and the same arrangement
// every other "one plain ability plus one costly ability" land has.
//
// No other simplification.
func init() {
	Register(Spec{
		OracleID:     "7fab8d65-af51-47d3-8f10-2676bf6e8ba3",
		Name:         "Eldrazi Temple",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}{C}",
				Label:    "Add {C}{C} (colorless Eldrazi only)",
				Restrictions: []string{
					ManaRestrictColorless,
					ManaRestrictSubtype("Eldrazi"),
				},
			},
		},
	})
}
