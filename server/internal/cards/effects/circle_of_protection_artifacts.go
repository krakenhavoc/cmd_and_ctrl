package effects

// Circle of Protection: Artifacts — Enchantment {1}{W}:
//
//	"{2}: The next time an artifact source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a chosen source (CR 615.8). The source is chosen as the
// ability resolves (CR 609.7a), from the permanents, spells, face-up
// command-zone cards and objects something refers to that are
// an artifact source. That is checked again when the source would deal the damage
// (CR 615.9): if it no longer matches, the damage is dealt and the
// shield is not used up.
//
// No simplifications.
func init() {
	Register(circleOfProtection("2e61e9c0-a2b9-4a24-8cb0-5160accce183", "Circle of Protection: Artifacts", "{2}",
		"{2}: The next time an artifact source of your choice would deal damage to you this turn, prevent that damage.",
		"an artifact source", QueryTypes("artifact")))
}
