package effects

// Auriok Transfixer — Creature — Human Scout {W}, 1/1 (Mirrodin, #5606):
//
//	"{W}, {T}: Tap target artifact."
//
// Arashin Sunshield's tapper on an artifact clause: any artifact, an
// artifact creature or an opponent's included, and Staff of Domination's
// body (tapChosenPermanent).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ad5f4c5-f6d2-4e96-af65-5488d9711113",
		Name:         "Auriok Transfixer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target artifact.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  tapChosenPermanent,
		}},
	})
}
