package effects

// Pilgrim of Justice — Creature — Human Cleric {2}{W}:
//
//	"Protection from red
//	 {W}, Sacrifice this creature: The next time a red source of your choice would deal damage this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// The source must be red to be offered and is checked again when it would
// deal the damage (CR 615.9). The shield protects nothing in particular:
// the chosen source's next damage is prevented, whoever it would be dealt
// to.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "959b672e-d4e4-49a1-874a-5148f040fadd",
		Name:            "Pilgrim of Justice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{W}, Sacrifice this creature: The next time a red source of your choice would deal damage this turn, prevent that damage.",
			Plus(ManaCost("{W}"), SacrificeThis()), nil,
			PreventNextDamageFromChosenSource(ShieldAnything, QueryColors("R")))},
	})
}
