package effects

// Galvanic Key — Artifact {2} (Mirrodin, #5606):
//
//	"Flash
//	 {3}, {T}: Untap target artifact."
//
// Voltaic Key at a higher price, with flash (CR 702.8a): it can be cast
// whenever an instant could. "Target artifact" admits the Key itself, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8a552ca-2c7b-410e-bd7e-1bb81465277a",
		Name:            "Galvanic Key",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Untap target artifact.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  untapFirstLegalTarget,
		}},
	})
}
