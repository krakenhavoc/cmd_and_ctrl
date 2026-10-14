package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omnipresence — Enchantment {5}{G}{G}{G}:
//
//	"You may cast spells with mana value less than or equal to the
//	 number of creatures you control from your hand without paying their
//	 mana costs."
//
// #3503, ADR 0118's amendment of 2026-10-14. Omniscience's granted free
// cast from hand with a cap that follows the board
// (GrantedAlternativeCost.MaxManaValueFunc): the creature count is read
// each time the offer is derived, so a creature entering or dying
// changes what may be cast for free at once. A spell keeps its own
// timing, X is 0 (CR 107.3b), and a card with no mana cost has mana
// value 0 and is covered when the cap allows it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "8dad1fa8-3533-4e3c-b9b3-fcd57b20a33d",
		Name:                    "Omnipresence",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{CastFromHandWithoutPayingManaCostUpToCreatureCount()},
	})
}
