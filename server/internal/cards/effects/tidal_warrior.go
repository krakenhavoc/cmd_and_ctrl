package effects

// Tidal Warrior — Creature — Merfolk Warrior {U}, 1/1:
//
//	"{T}: Target land becomes an Island until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved ability. Until end of
// turn the land's land types are replaced by Island (its other subtypes
// stay, CR 205.1a), it loses the abilities its rules text gives it, keeps
// any another effect granted it, and taps for {U} (CR 305.6). A land
// that has become an illegal target by resolution is left alone (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a1bc869-08a1-4a70-9c9f-2e02e47cd95b",
		Name:         "Tidal Warrior",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target land becomes an Island until end of turn.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Tidal Warrior", "Island"),
		}},
	})
}
