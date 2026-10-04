package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// BuybackDiscard is a NON-MANA buyback that discards cards — "Buyback—
// Discard two cards" (Demonic Collusion). Its siblings are Buyback
// (mana) and BuybackSacrifice (a permanent).
//
// The discards are paid with the spell already on the stack
// (CR 601.2h), like any additional-cost discard, and the return to
// hand is the engine's (CR 702.27a), so a card using this declares the
// cost and nothing else.
func BuybackDiscard(n int) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:     true,
		Key:          game.BuybackKey,
		DiscardCards: n,
		Label:        "Buyback—" + discardLabel(n),
	}
}
