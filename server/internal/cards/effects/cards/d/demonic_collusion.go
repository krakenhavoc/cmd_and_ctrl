package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "53cbc12d-e182-4cde-8967-9b2664daa817",
		Name:          "Demonic Collusion",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.BuybackDiscard(2)},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
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
