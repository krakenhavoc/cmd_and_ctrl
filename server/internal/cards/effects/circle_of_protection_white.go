package effects

// Circle of Protection: White — Enchantment {1}{W}:
//
//	"{1}: The next time a white source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a chosen source (CR 615.8). The source is chosen as the
// ability resolves (CR 609.7a), from the permanents, spells, face-up
// command-zone cards and objects something refers to that are
// a white source. That is checked again when the source would deal the damage
// (CR 615.9): if it no longer matches, the damage is dealt and the
// shield is not used up.
//
// No simplifications.
func init() {
	Register(circleOfProtection("5f46f86a-9779-4ed3-99cb-76a03d380598", "Circle of Protection: White", "{1}",
		"{1}: The next time a white source of your choice would deal damage to you this turn, prevent that damage.",
		"a white source", QueryColors("W")))
}
