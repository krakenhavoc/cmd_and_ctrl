package effects

// Rowan, Scion of War — Legendary Creature — Human Wizard {1}{B}{R}, 3/3
// (Wilds of Eldraine, #4471):
//
//	"Menace
//	 {T}: Spells you cast this turn that are black and/or red cost {X}
//	 less to cast, where X is the amount of life you lost this turn.
//	 Activate only as a sorcery."
//
// The discount is Goblin Maskmaker's every-matching-spell promise with
// the amount fixed as the ability resolves from the turn's life-lost
// tally (scionDiscountAbility): it counts the total lost, ignoring life
// gained, and cannot reduce coloured mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "333c15c6-c8fd-4cdd-89d2-06b921edeb8b",
		Name:            "Rowan, Scion of War",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{scionDiscountAbility(
			"{T}: Spells you cast this turn that are black and/or red cost {X} less to cast, where X is the amount of life you lost this turn. Activate only as a sorcery.",
			"Rowan, Scion of War", "BR", false)},
	})
}
