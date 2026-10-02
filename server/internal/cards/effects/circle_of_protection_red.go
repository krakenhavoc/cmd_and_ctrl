package effects

// Circle of Protection: Red — Enchantment {1}{W}:
//
//	"{1}: The next time a red source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a chosen source (CR 615.8). The source is chosen as the
// ability resolves (CR 609.7a), from the permanents, spells, face-up
// command-zone cards and objects something refers to that are
// a red source. That is checked again when the source would deal the damage
// (CR 615.9): if it no longer matches, the damage is dealt and the
// shield is not used up.
//
// No simplifications.
func init() {
	Register(circleOfProtection("df2738fe-9cd1-4347-8808-105fcfde1190", "Circle of Protection: Red", "{1}",
		"{1}: The next time a red source of your choice would deal damage to you this turn, prevent that damage.",
		"a red source", QueryColors("R")))
}
