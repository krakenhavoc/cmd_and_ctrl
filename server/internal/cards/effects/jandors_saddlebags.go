package effects

// Jandor's Saddlebags — Artifact {2} (Arabian Nights, #5641):
//
//	"{3}, {T}: Untap target creature."
//
// Thousand-Year Elixir's untap row at {3}: any creature, whoever
// controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3aa0e73f-ac88-47a5-9cc5-0c941a939eae",
		Name:         "Jandor's Saddlebags",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Untap target creature.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  untapFirstLegalTarget,
		}},
	})
}
