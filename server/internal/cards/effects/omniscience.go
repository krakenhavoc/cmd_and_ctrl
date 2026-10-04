package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omniscience — Enchantment {7}{U}{U}{U}:
//
//	"You may cast spells from your hand without paying their mana costs."
//
// The granted-alternative-cost seam at {0} (ADR 0118 §3, #2163), for the
// hand only. A spell keeps its own timing, so a sorcery still waits for
// a main phase with an empty stack; X is 0 (CR 107.3b); a card with no
// mana cost can be cast this way (CR 118.6a).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                "730e39e6-c61d-48b5-8827-bfd952bf1be7",
		Name:                    "Omniscience",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{CastFromHandWithoutPayingManaCost()},
	})
}
