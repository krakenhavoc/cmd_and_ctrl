package effects

// Penance — Enchantment {2}{W}:
//
//	"Put a card from your hand on top of your library: The next time a
//	 black or red source of your choice would deal damage this turn,
//	 prevent that damage."
//
// ADR 0109 §7 (#1902): "Put a card from your hand on top of your
// library" is a cost the activator names at announce (CR 602.2b), never
// the card being activated. It is not a discard, and the activator still
// knows the card they put there.
// The shield is ADR 0107 §6's: the next instance of damage from a source
// chosen as the ability resolves (CR 615.8, 609.7a), to anything. The
// source must be black or red to be offered and is checked again when
// it would deal the damage (CR 615.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7818d6e0-74d8-46ae-94f6-d2a349c9507b",
		Name:         "Penance",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"Put a card from your hand on top of your library: The next time a black or red source of your choice would deal damage this turn, prevent that damage.",
			PutACardFromHandOnTop(), nil,
			PreventNextDamageFromChosenSource(ShieldAnything, QueryColors("B", "R")))},
	})
}
