package effects

// Ali Baba — Creature — Human Rogue {R}, 1/1 (Arabian Nights, #5641):
//
//	"{R}: Tap target Wall."
//
// Any Wall, whoever controls it. No {T} in the cost, so Ali Baba can use
// it the turn he arrives, and as often as there is red mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef3fdfea-330d-4948-94ec-67e344b48086",
		Name:         "Ali Baba",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: Tap target Wall.",
			Cost:    ManaCost("{R}"),
			Targets: TargetPermanent("target Wall", HasSubtype("Wall")),
			Effect:  tapTheFirstLegalTarget,
		}},
	})
}
