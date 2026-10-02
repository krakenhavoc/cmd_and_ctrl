package effects

// Cho-Arrim Alchemist — Creature — Human Spellshaper {W}:
//
//	"{1}{W}{W}, {T}, Discard a card: The next time a source of your choice would deal damage to you this turn, prevent that damage. You gain life equal to the damage prevented this way."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "You gain life equal to the damage prevented this way" is the shield's
// own additional effect, run immediately after the prevention (CR 615.5,
// owner decision 4).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1e1d4d81-4a45-41b5-a781-ae2227b4f4c0",
		Name:         "Cho-Arrim Alchemist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{1}{W}{W}, {T}, Discard a card: The next time a source of your choice would deal damage to you this turn, prevent that damage. You gain life equal to the damage prevented this way.",
			Plus(ManaCost("{1}{W}{W}"), TapCost(), DiscardACard()), nil,
			PreventNextDamageFromChosenSource(ShieldYou).WithThen(preventedGainLifeBody))},
	})
}
