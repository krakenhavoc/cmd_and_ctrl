package effects

// Mystic Compass — Artifact {2}:
//
//	"{1}, {T}: Target land becomes the basic land type of your choice until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved ability. The basic land
// type is chosen as the ability resolves (CR 608.2), by its controller,
// through effects.ChooseBasicLandTypeThen; a land that has become an
// illegal target by then is left alone and nothing is asked (CR 608.2b).
// Until end of turn the land's land types are replaced by the chosen one
// (its other subtypes stay, CR 205.1a), it loses the abilities its rules
// text gives it, keeps any another effect granted it, and taps for the
// chosen type's colour (CR 305.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87bafa97-38fd-41a2-8b1d-315e88d9d0bf",
		Name:         "Mystic Compass",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target land becomes the basic land type of your choice until end of turn.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Mystic Compass"),
		}},
	})
}
