package effects

// King Suleiman — Creature — Human Noble {1}{W}, 1/1 (Arabian Nights,
// #5641):
//
//	"{T}: Destroy target Djinn or Efreet."
//
// Royal Assassin's row with a subtype clause: any permanent that is a
// Djinn or an Efreet, whoever controls it, read off its current types
// (a changeling counts).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97f548b1-6398-4dc7-b5eb-da5a3f24ddb7",
		Name:         "King Suleiman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Destroy target Djinn or Efreet.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target Djinn or Efreet", Or(HasSubtype("Djinn"), HasSubtype("Efreet"))),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
