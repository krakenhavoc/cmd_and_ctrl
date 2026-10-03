package effects

// Burrenton Forge-Tender — Creature — Kithkin Wizard {W}:
//
//	"Protection from red
//	 Sacrifice this creature: Prevent all damage a red source of your choice would deal this turn."
//
// ADR 0108 §7 (#1904): the source must be red to be offered, and is
// checked again each time it would deal damage this turn (CR 615.9): a
// source that stops being red gets through, and is prevented again if it
// turns red. The shield protects nothing in particular — all of the
// chosen source's damage is prevented, whoever it would be dealt to.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "70bb275b-3458-4690-a50f-b231fbf0bccb",
		Name:            "Burrenton Forge-Tender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Activated: []ActivatedAbility{sourceShieldRow(
			"Sacrifice this creature: Prevent all damage a red source of your choice would deal this turn.",
			SacrificeThis(), nil,
			PreventDamageFromChosenSource(ShieldAnything, QueryColors("R")))},
	})
}
