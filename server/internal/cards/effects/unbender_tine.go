package effects

// Unbender Tine — Artifact {2}{W}{U} (Alara Reborn, #5586):
//
//	"{T}: Untap another target permanent."
//
// Formidable Speaker's untap row without its mana: "another" is object
// identity (CR 109.1), so the Tine can't target itself, and a target that
// left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdc0a1a5-d9e4-4923-b472-2fa78212167f",
		Name:         "Unbender Tine",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Untap another target permanent.",
			Cost:    TapCost(),
			Targets: Another(TargetPermanent("another target permanent", Permanent())),
			Effect:  b22UntapEachLegalTarget,
		}},
	})
}
