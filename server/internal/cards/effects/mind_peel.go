package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Peel — Sorcery {B} (Stronghold, #5627):
//
//	"Buyback {2}{B}{B} (You may pay an additional {2}{B}{B} as you cast
//	 this spell. If you do, put this card into your hand as it resolves.)
//	 Target player discards a card."
//
// Raven's Crime's discard, through the shared targetOpponentDiscards
// body (it queues the discard prompt for whichever player the item
// targets, so "target player" includes the caster), with Capsize's
// buyback (CR 702.27a). The discarding player picks the card
// (CR 701.9b); an empty hand discards nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "4f424d94-1544-4086-a57d-2add17332094",
		Name:          "Mind Peel",
		Completeness:  CompletenessFull,
		Targets:       TargetPlayer("target player"),
		OptionalCosts: []game.AdditionalCost{Buyback("{2}{B}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetOpponentDiscards(1)(ctx.Game, item)
		},
	})
}
