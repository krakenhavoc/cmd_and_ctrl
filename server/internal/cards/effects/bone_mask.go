package effects

// Bone Mask — Artifact {4}:
//
//	"{2}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage. Exile cards from the top of your library equal to the damage prevented this way."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "Exile cards from the top of your library equal to the damage
// prevented this way" is the shield's own additional effect, run
// immediately after the prevention with the amount prevented (CR 615.5,
// owner decision 4); after damage that can't be prevented it exiles
// nothing (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ddfc7dac-f921-4529-935c-6c588890c522",
		Name:         "Bone Mask",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{2}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage. Exile cards from the top of your library equal to the damage prevented this way.",
			Plus(ManaCost("{2}"), TapCost()), nil,
			PreventNextDamageFromChosenSource(ShieldYou).WithThen(preventedExileLibraryTopBody))},
	})
}
