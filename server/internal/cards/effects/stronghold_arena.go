package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stronghold Arena — Enchantment {1}{B}:
//
//	"Kicker {G} and/or {W} (You may pay an additional {G} and/or {W}
//	 as you cast this spell.)
//	 When this enchantment enters, you gain 3 life for each time it was
//	 kicked.
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, you may reveal the top card of your library and put it
//	 into your hand. If you do, you lose life equal to its mana value."
//
// Two kicker costs (CR 702.33b, #2153) that are not linked to separate
// abilities: the enters trigger counts them together (CR 702.33d), so a
// doubly kicked Arena gains six. The count is read once in Build, off
// the record the resolution carried onto the permanent (CR 400.7d), and
// stamped on the item; the trigger fires unkicked too and gains nothing,
// which is the printed text (no intervening if).
//
// The combat trigger is the "one or more" shape, once per player
// connected with (CR 603.2c). The flip is a reveal, not a draw, so a
// draw payoff does not see it (revealTopReadingManaValue).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "7dabd64c-8d0b-4e56-b6be-5651c9e985c5",
		Name:          "Stronghold Arena",
		Completeness:  CompletenessFull,
		OptionalCosts: Kickers("{G}", "{W}"),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Stronghold Arena — gain 3 life for each time it was kicked",
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Stronghold Arena — gain 3 life for each time it was kicked")
					item.Params.Amount = 3 * game.CardKickedTimes(*source)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Amount: item.Params.Amount}.Apply(NewContext(g, item))
				},
			},
			Optional(WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Stronghold Arena — reveal the top card of your library and put it into your hand",
				strongholdArenaFlip), "Reveal the top card of your library and put it into your hand? You lose life equal to its mana value."),
		},
	})
}

// strongholdArenaFlip is the optional flip: the top card is revealed
// and put into the controller's hand, and "if you do" they lose life
// equal to its mana value. The loss is gated on the card actually
// reaching a hand (BounceToHand.Then), so a move the CR 614 window
// cancelled or redirected costs nothing.
func strongholdArenaFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	flipped, mv, err := revealTopReadingManaValue(ctx, "Stronghold Arena — reveal the top card of your library")
	if err != nil || flipped == uuid.Nil {
		return err
	}
	return BounceToHand{Target: flipped, Then: func(ctx *Context, bounced bool) error {
		if !bounced || mv == 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -mv)
	}}.Apply(ctx)
}
