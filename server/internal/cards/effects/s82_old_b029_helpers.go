package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b029_helpers.go — shared pieces of OLD batch 29 (#5501, Modern
// Horizons 2).

// OldB029DrawTwoAfterDiscard is the "then draw two cards" of "Discard a
// card, then draw two cards" (Fast). The draw is a registered
// continuation, so a table waiting on the discard is a restore point. It
// runs even when the hand was empty and nothing was discarded (CR 701.9a).
var OldB029DrawTwoAfterDiscard = game.RegisterDiscardThen("choose-cards/old-b029-fast-draw-two",
	func(g *game.Game, r game.DiscardsDone) error {
		return g.DrawNForEffect(r.Player, 2)
	})
