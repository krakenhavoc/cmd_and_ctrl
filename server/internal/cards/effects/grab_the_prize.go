package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grab the Prize — Sorcery {1}{R}:
//
//	"As an additional cost to cast this spell, discard a card. Draw two
//	 cards. If the discarded card wasn't a land card, Grab the Prize
//	 deals 2 damage to each opponent."
//
// The discard is an ordinary additional cost (ADR 0021); "the discarded
// card" is read off the payment record, PaidCost.Discarded (ADR 0100,
// owner decision 6), because by resolution nothing on the board says
// which card paid. The card keeps its instance ID wherever it has gone
// since — the graveyard, or exile under madness — and whether it is a
// land card is a printed characteristic, so the lookup answers for it
// wherever it is. A card that cannot be found at all is the unknown
// answer, and deals nothing.
//
// A copy asks about the card the original discarded (CR 707.10: "it
// uses the objects used to pay the costs of the original").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "a37387f0-9d57-49ff-aa38-2c6b359e8916",
		Name:           "Grab the Prize",
		Completeness:   CompletenessFull,
		AdditionalCost: DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			discarded := ctx.Discarded()
			if len(discarded) == 0 {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(discarded[0])
			if !ok || card.IsLand() {
				return nil
			}
			return damageToEachOpponent(ctx.Game, item, 2)
		},
	})
}
