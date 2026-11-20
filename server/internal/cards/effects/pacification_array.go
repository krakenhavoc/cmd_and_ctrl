package effects

// Pacification Array — Artifact {1} (Aether Revolt,
// #5536):
//
//	"{2}, {T}: Tap target artifact or creature."
//
// Sterling Keykeeper's tap row with an artifact-or-creature target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2c14195-2747-490e-99ef-5b71048ff51a",
		Name:         "Pacification Array",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Tap target artifact or creature.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
			Effect:  tapChosenPermanent,
		}},
	})
}
