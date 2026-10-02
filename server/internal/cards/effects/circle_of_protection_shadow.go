package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Circle of Protection: Shadow — Enchantment {1}{W}:
//
//	"{1}: The next time a creature of your choice with shadow would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a chosen source (CR 615.8). The source is chosen as the
// ability resolves (CR 609.7a), from the permanents, spells, face-up
// command-zone cards and objects something refers to that are
// a creature with shadow. That is checked again when the source would deal the damage
// (CR 615.9): if it no longer matches, the damage is dealt and the
// shield is not used up.
//
// No simplifications.
func init() {
	Register(circleOfProtection("07c42f46-800e-4d2a-b844-c125ab93190d", "Circle of Protection: Shadow", "{1}",
		"{1}: The next time a creature of your choice with shadow would deal damage to you this turn, prevent that damage.",
		"a creature with shadow", game.PermanentQuery{Types: []string{"creature"}, Keyword: "shadow"}))
}
