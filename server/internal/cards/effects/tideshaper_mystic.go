package effects

// Tideshaper Mystic — Creature — Merfolk Wizard {U}, 1/1:
//
//	"{T}: Target land becomes the basic land type of your choice until end of turn. Activate only during your turn."
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
// "Activate only during your turn" is the activation condition: any step
// of your turn, not only a sorcery-speed window.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3eec0ca6-e461-4583-9989-e4f6b15a53f8",
		Name:         "Tideshaper Mystic",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target land becomes the basic land type of your choice until end of turn. Activate only during your turn.",
			Cost:      TapCost(),
			Targets:   TargetPermanent("target land", Land()),
			Condition: DuringYourTurn(),
			Effect:    TargetLandBecomesUntilEOT("Tideshaper Mystic"),
		}},
	})
}
