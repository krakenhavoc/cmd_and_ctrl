package effects

// Barbarian Ring — Land (EDHREC rank 4792):
//
//	"{T}: Add {R}. This land deals 1 damage to you.
//	 Threshold — {R}, {T}, Sacrifice this land: It deals 2 damage to
//	 any target. Activate only if there are seven or more cards in
//	 your graveyard."
//
// Odyssey's red threshold land. The mana ability is a painland half:
// the damage is a rider, not a cost. Threshold is the second ability's
// activation condition (CR 602.1b, #743), GraveyardAtLeast(7) over any
// cards. The land is sacrificed as a cost, so "it" deals the damage
// from the graveyard, by last known information — still a red source
// of nothing but a land, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eeb9377b-72c1-4214-9a66-0f55577c17d1",
		Name:         "Barbarian Ring",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}. This land deals 1 damage to you.",
			Rider:    PainRider(1),
		}},
		Activated: []ActivatedAbility{{
			Label:     "Threshold — {R}, {T}, Sacrifice this land: It deals 2 damage to any target. Activate only if there are seven or more cards in your graveyard.",
			Cost:      Plus(ManaCost("{R}"), TapCost(), SacrificeThis()),
			Targets:   TargetAny(),
			Condition: GraveyardAtLeast(7, nil),
			Effect:    b33DamageChosenTargetFromSource(2),
		}},
	})
}
