package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightmare Void — Sorcery {3}{B}:
//
//	"Target player reveals their hand. You choose a card from it. That
//	 player discards that card.
//	 Dredge 2 (If you would draw a card, you may mill two cards
//	 instead. If you do, return this card from your graveyard to your
//	 hand.)"
//
// Coercion's revealed-hand pick (ADR 0116) over ANY player, not just an
// opponent, plus dredge 2 (dredge.go, #2127). Because the card is
// discarded to its owner's graveyard by a spell that then lands in the
// same graveyard, a Void that dredges back is a recurring discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a53ef0d-8e29-4128-9cc4-fbdd3e4fdb84",
		Name:         "Nightmare Void",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Replacements: []game.ReplacementEffect{Dredge(2)},
		OnResolve:    TargetRevealsYouChooseDiscard(nil, "card"),
	})
}
