package effects

// Circle of Protection: Blue — Enchantment {1}{W}:
//
//	"{1}: The next time a blue source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a chosen source (CR 615.8). The source is chosen as the
// ability resolves (CR 609.7a), from the permanents, spells, face-up
// command-zone cards and objects something refers to that are
// a blue source. That is checked again when the source would deal the damage
// (CR 615.9): if it no longer matches, the damage is dealt and the
// shield is not used up.
//
// No simplifications.
func init() {
	Register(circleOfProtection("d7572f17-f85d-45c0-ac64-43aac760eafe", "Circle of Protection: Blue", "{1}",
		"{1}: The next time a blue source of your choice would deal damage to you this turn, prevent that damage.",
		"a blue source", QueryColors("U")))
}
