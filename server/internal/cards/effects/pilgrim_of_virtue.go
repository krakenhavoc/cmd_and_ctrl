package effects

// Pilgrim of Virtue — Creature — Human Cleric {2}{W}:
//
//	"Protection from black
//	 {W}, Sacrifice this creature: The next time a black source of your choice would deal damage this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// The source must be black to be offered and is checked again when it would
// deal the damage (CR 615.9). The shield protects nothing in particular:
// the chosen source's next damage is prevented, whoever it would be dealt
// to.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "e2f73743-95e7-4f03-9b98-5ab5ebc4b724",
		Name:            "Pilgrim of Virtue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{W}, Sacrifice this creature: The next time a black source of your choice would deal damage this turn, prevent that damage.",
			Plus(ManaCost("{W}"), SacrificeThis()), nil,
			PreventNextDamageFromChosenSource(ShieldAnything, QueryColors("B")))},
	})
}
