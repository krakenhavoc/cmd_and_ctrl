package effects

// Shrine of the Forsaken Gods — Land:
//
//	"{T}: Add {C}."
//	"{T}: Add {C}{C}. Spend this mana only to cast colorless spells.
//	 Activate only if you control seven or more lands."
//
// Rank 1226, and the card that exercises two of #352's seams on one
// ability: a spend restriction AND an activation gate. Both are
// drawbacks, both had to be real before the card could ship, and
// either one missing would make this a land that is better than
// printed — the second ability with neither is a free Ancient Tomb.
//
// The gate is checked first, before any cost (CR 602.5), so an
// activation at six lands fails with the land untapped. The
// restriction is stamped onto the tokens and enforced by the spend
// solver when a spell's cost is paid.
//
// "Colorless SPELLS", not "colorless Eldrazi" — so unlike Eldrazi
// Temple this one names the kind of payment (casting, not activating)
// and one property of the object (colorless), and no subtype.
// ManaRestrictCast plus ManaRestrictColorless is the whole clause.
//
// Like Temple of the False God, the land counts itself among the
// seven.
//
// The plain "{T}: Add {C}" is unrestricted and ungated and auto-taps
// normally. The big ability does not — a gated, restricted source is
// two decisions the planner cannot make.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ea46945-d5ab-4209-b473-4769e7b8b962",
		Name:         "Shrine of the Forsaken Gods",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:      ManaAbilityCost{Tap: true},
				Produced:  "{C}{C}",
				Label:     "Add {C}{C} (colorless spells only, seven+ lands)",
				Condition: ControlsAtLeast(7, MatchLand),
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictColorless,
				},
			},
		},
	})
}
