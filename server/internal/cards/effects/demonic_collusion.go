package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demonic Collusion — Sorcery {3}{B}{B}:
//
//	"Buyback—Discard two cards. (You may discard two cards in addition
//	 to any other costs as you cast this spell. If you do, put this
//	 card into your hand as it resolves.)
//	 Search your library for a card, put that card into your hand, then
//	 shuffle."
//
// Demonic Tutor's body behind a non-mana buyback (BuybackDiscard). The
// two discards are paid as the spell goes on the stack; the return to
// hand is the engine's, not this file's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "53cbc12d-e182-4cde-8967-9b2664daa817",
		Name:          "Demonic Collusion",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{BuybackDiscard(2)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  false,
				Shuffle: true,
				Reason:  "Demonic Collusion — search your library for a card",
			}.Apply(ctx)
		},
	})
}
