package effects

// Rune of Protection: Green — Enchantment {1}{W}:
//
//	"{W}: The next time a green source of your choice would deal damage to you this turn, prevent that damage.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a chosen source (CR 615.8). The source is chosen as the
// ability resolves (CR 609.7a), from the permanents, spells, face-up
// command-zone cards and objects something refers to that are
// a green source. That is checked again when the source would deal the damage
// (CR 615.9): if it no longer matches, the damage is dealt and the
// shield is not used up.
//
// No simplifications.
func init() {
	Register(runeOfProtection("c7fca3f3-b48e-4ac5-be31-ce7ae3c7432c", "Rune of Protection: Green",
		"{W}: The next time a green source of your choice would deal damage to you this turn, prevent that damage.",
		"a green source", QueryColors("G")))
}
