package effects

// Rune of Protection: Red — Enchantment {1}{W}:
//
//	"{W}: The next time a red source of your choice would deal damage to you this turn, prevent that damage.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
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
	Register(runeOfProtection("9b3f6b29-675a-4153-9582-03318851b20e", "Rune of Protection: Red",
		"{W}: The next time a red source of your choice would deal damage to you this turn, prevent that damage.",
		"a red source", QueryColors("R")))
}
