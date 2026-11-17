package effects

// Southern Paladin — Creature — Human Knight {2}{W}{W}, 3/3 (Weatherlight, #5629):
//
//	"{W}{W}, {T}: Destroy target red permanent."
//
// The target's colour is read at activation and again at resolution
// (CR 608.2b), so a permanent that stopped being red is not destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cab96aa8-13e3-4dd4-9d44-9e298f807388",
		Name:         "Southern Paladin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}{W}, {T}: Destroy target red permanent.",
			Cost:    Plus(ManaCost("{W}{W}"), TapCost()),
			Targets: TargetPermanent("target red permanent", OfColor("R")),
			Effect:  destroyChosenPermanent,
		}},
	})
}
