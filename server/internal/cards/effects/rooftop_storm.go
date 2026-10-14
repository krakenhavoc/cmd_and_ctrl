package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rooftop Storm — Enchantment {5}{U}:
//
//	"You may pay {0} rather than pay the mana cost for Zombie creature
//	 spells you cast."
//
// A granted alternative cost (ADR 0118 §3): every Zombie creature spell
// its controller casts is offered one more way to pay, from any zone
// the spell may be cast from (CR 118.9a), and X is 0 (CR 107.3b).
// Additional costs are still paid. Only one alternative cost may be
// applied to a spell (CR 118.9), so it does not combine with another.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "6835bf6d-7197-481b-ac26-276ce363b4ad",
		Name:                    "Rooftop Storm",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayZeroForZombieCreatureSpellsYouCast()},
	})
}
