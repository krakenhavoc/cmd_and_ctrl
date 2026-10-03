package effects

// Grixis Illusionist — Creature — Human Wizard {U}, 1/1:
//
//	"{T}: Target land you control becomes the basic land type of your choice until end of turn."
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
		OracleID:     "c231698b-c584-4047-9668-5322e4b0065b",
		Name:         "Grixis Illusionist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target land you control becomes the basic land type of your choice until end of turn.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land you control", Land(), YouControl()),
			Effect:  TargetLandBecomesUntilEOT("Grixis Illusionist"),
		}},
	})
}
