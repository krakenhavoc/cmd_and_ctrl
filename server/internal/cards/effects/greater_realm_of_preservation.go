package effects

// Greater Realm of Preservation — Enchantment {1}{W}:
//
//	"{1}{W}: The next time a black or red source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// The source must be black or red to be offered, and is checked again
// when it would deal the damage (CR 615.9).
//
// No simplifications.
func init() {
	Register(circleOfProtection("b03eb0c4-89a4-420d-8a23-1a868a07d9cf", "Greater Realm of Preservation", "{1}{W}",
		"{1}{W}: The next time a black or red source of your choice would deal damage to you this turn, prevent that damage.",
		"a black or red source", QueryColors("B", "R")))
}
