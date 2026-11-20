package effects

// Blinding Mage — Creature — Human Wizard {1}{W}, 1/2 (Magic 2010, #5585):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tap ability: the target is chosen on activation and
// tapped as it resolves, if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53a1faa1-ca67-4027-bee8-0f7c4584d62b",
		Name:         "Blinding Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
