package effects

// Will, Scion of Peace — Legendary Creature — Human Wizard {1}{W}{U}, 2/4
// (Wilds of Eldraine, #4471):
//
//	"Vigilance
//	 {T}: Spells you cast this turn that are white and/or blue cost {X}
//	 less to cast, where X is the amount of life you gained this turn.
//	 Activate only as a sorcery."
//
// Rowan's discount over the turn's life-gained tally
// (scionDiscountAbility), which ignores life lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4ecb8653-db8c-468a-853d-420ff62a7a8a",
		Name:            "Will, Scion of Peace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{scionDiscountAbility(
			"{T}: Spells you cast this turn that are white and/or blue cost {X} less to cast, where X is the amount of life you gained this turn. Activate only as a sorcery.",
			"Will, Scion of Peace", "WU", true)},
	})
}
