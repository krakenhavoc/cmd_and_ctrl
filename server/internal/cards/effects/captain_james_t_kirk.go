package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captain James T. Kirk — {2}{R} Legendary Creature — Human Officer, 2/2
// (Star Trek, #4973):
//
//	"Whenever Captain Kirk enters or attacks, choose one. If you have no
//	 cards in hand, choose one or more instead.
//	 • Discard a card, then draw a card.
//	 • Create a 1/1 red Officer creature token.
//	 • Creatures you control get +1/+0 until end of turn."
//
// One trigger, two events (WhenThisEntersOrAttacks, Sun Titan's shape).
// The mode count is Let's Play a Game's "choose one or more instead"
// (AnyNumberIf) over a hand-size condition, read for the controller as the
// trigger is put on the stack, which is when the modes are chosen (CR
// 603.3c). The rummage is Charming Scoundrel's: discard first, then draw
// even from an empty hand. The +1/+0 is a one-shot set fixed as it
// resolves (CR 611.2c), so a creature that arrives later does not get it.
//
// No simplification.
func init() {
	kirk := WhenThisEntersOrAttacks("Captain James T. Kirk — choose one (one or more with no cards in hand)",
		func(*game.Game, *game.StackItem) error { return nil })
	kirk.Modes = ChooseOne(
		ModeDoing("Discard a card, then draw a card.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: item.Controller,
				Source: item.SourceCardID,
				N:      1,
			}, captainKirkDraw, nil)
		}),
		ModeDoing("Create a 1/1 red Officer creature token.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 red Officer"), N: 1}.Apply(ctx)
		}),
		ModeDoing("Creatures you control get +1/+0 until end of turn.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl()),
				Power: 1,
				Label: "Captain James T. Kirk — +1/+0 until end of turn",
			}.Apply(ctx)
		}),
	).AnyNumberIf(YouHaveNoCardsInHand)
	Register(Spec{
		OracleID:     "19edde54-153f-46b8-8a8d-a20fdef0667f",
		Name:         "Captain James T. Kirk",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{kirk},
	})
}

// YouHaveNoCardsInHand is "If you have no cards in hand, choose one or more
// instead" (Captain James T. Kirk): the chooser's hand, read as the modes
// are chosen.
var YouHaveNoCardsInHand = game.ModeCondition("you-have-no-cards-in-hand", func(g *game.Game, chooser uuid.UUID) bool {
	p := g.PlayerByIDForEffect(chooser)
	return p != nil && (p.Hand == nil || p.Hand.Size() == 0)
})

// captainKirkDraw is the rummage's "then draw a card", keyed: a table
// waiting on the discard is a restore point.
var captainKirkDraw = game.RegisterDiscardThen("choose-cards/captain-james-t-kirk-draw", func(g *game.Game, r game.DiscardsDone) error {
	return g.DrawNForEffect(r.Player, 1)
})
