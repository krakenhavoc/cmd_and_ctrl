package effects

// Circle of Protection: Black — Enchantment {1}{W}:
//
//	"{1}: The next time a black source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a chosen source (CR 615.8). The source is chosen as the
// ability resolves (CR 609.7a), from the permanents, spells, face-up
// command-zone cards and objects something refers to that are
// a black source. That is checked again when the source would deal the damage
// (CR 615.9): if it no longer matches, the damage is dealt and the
// shield is not used up.
//
// No simplifications.
func init() {
	Register(circleOfProtection("7a5a8414-4da4-4dd0-93ae-210d50f4d6f6", "Circle of Protection: Black", "{1}",
		"{1}: The next time a black source of your choice would deal damage to you this turn, prevent that damage.",
		"a black source", QueryColors("B")))
}
