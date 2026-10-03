package effects

// Reef Shaman — Creature — Merfolk Shaman {U}, 0/2:
//
//	"{T}: Target land becomes the basic land type of your choice until end of turn."
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
		OracleID:     "4ad870b2-133c-4229-84a8-f91e4c62fcd0",
		Name:         "Reef Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target land becomes the basic land type of your choice until end of turn.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Reef Shaman"),
		}},
	})
}
