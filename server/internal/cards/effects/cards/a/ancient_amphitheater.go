package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Amphitheater — Land (Lorwyn, #5593):
//
//	"As this land enters, you may reveal a Giant card from your hand. If you
//	 don't, this land enters tapped.
//	 {T}: Add {R} or {W}."
//
// Secluded Glen's reveal-land entry over Giant cards: the reveal is a
// choice the player may decline (CR 614.12), the card stays in hand
// (CR 701.20b), and the test is the type line of a card in hand, a
// Kindred card or a changeling card counting as a Giant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7211221d-d4d8-4bbe-9d2a-b82e005bfe8a",
		Name:         "Ancient Amphitheater",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersTappedUnlessYouRevealFromHand("Ancient Amphitheater", "a Giant card", effects.LrwB1CardOfType("Giant")),
		},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("R", "W")},
	})
}
