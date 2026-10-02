package effects

// Tundra Kavu — Creature — Kavu {2}{R}, 2/2:
//
//	"{T}: Target land becomes a Plains or an Island until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved ability, with the choice
// narrowed to two. The controller chooses Plains or Island as the ability
// resolves (CR 608.2), through effects.ChooseBasicLandTypeThen; a land that
// has become an illegal target by then is left alone and nothing is asked
// (CR 608.2b). Until end of turn the land's land types are replaced by the
// chosen one (its other subtypes stay, CR 205.1a), it loses the abilities
// its rules text gives it, and it taps for {W} or {U} (CR 305.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0b7d88b-694b-4bbe-888d-05245e7d1c3e",
		Name:         "Tundra Kavu",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target land becomes a Plains or an Island until end of turn.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Tundra Kavu", "Plains", "Island"),
		}},
	})
}
