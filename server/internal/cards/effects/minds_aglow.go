package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minds Aglow — Sorcery {U} (Commander Masters, #5948):
//
//	"Join forces — Starting with you, each player may pay any amount of
//	 mana. Each player draws X cards, where X is the total amount of mana
//	 paid this way."
//
// Karn, Living Legacy's "pay any amount of mana" (MayPayX over {1}), asked
// of each living seat in turn order starting with the caster, each paying
// from their own mana (CR 608.2d). The running total rides the prompt as
// data: the seats still to ask first, then one placeholder (uuid.Nil) per
// mana paid so far. When the last seat has answered, every player draws the
// total; a seat that paid nothing or declined adds none, and a total of
// zero draws nothing. The draws are one instruction after the payments, in
// turn order from the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b259f4e4-04a9-40e0-85e9-34b4b74ee388",
		Name:         "Minds Aglow",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return mindsAglowAsk(ctx, seatsFromController(ctx.Game, item.Controller), 0)
		},
	})
}

// mindsAglowAsk asks seats[0] how much to pay, then carries on with the rest
// and the new total. A seat that cannot be asked adds nothing.
func mindsAglowAsk(ctx *Context, seats []uuid.UUID, total int) error {
	if len(seats) == 0 {
		return mindsAglowDraw(ctx, total)
	}
	carry := append([]uuid.UUID(nil), seats[1:]...)
	carry = append(carry, uuid.Nil) // the marker between the seats and the total
	for i := 0; i < total; i++ {
		carry = append(carry, uuid.Nil)
	}
	return MayPayX{
		Chooser: seats[0],
		Symbol:  "{1}",
		Label:   "Minds Aglow",
		Buys:    "each player draws X cards",
		Unit:    game.PayAmountCards,
		Then:    mindsAglowPaid,
		Carry:   carry,
	}.Apply(ctx)
}

// mindsAglowPaid is the join-forces step once a seat has settled its amount:
// the pay branch adds X, the decline branch adds nothing, and either asks
// the next seat. An on-disk key.
// Registered in init because its branches ask the next seat with it.
var mindsAglowPaid game.PayUnlessElse

func init() {
	mindsAglowPaid = PayUnlessBranches("pay-unless/minds-aglow-paid",
		func(ctx *Context, r game.PayUnlessAnswer) error { return mindsAglowNext(ctx, r, r.Number) },
		func(ctx *Context, r game.PayUnlessAnswer) error { return mindsAglowNext(ctx, r, 0) })
}

// mindsAglowNext reads the seats still to ask and the total so far off the
// carry (the seats, a marker, then one placeholder per mana paid)
// and asks the next seat with `paid` added.
func mindsAglowNext(ctx *Context, r game.PayUnlessAnswer, paid int) error {
	var seats []uuid.UUID
	total := 0
	pastMarker := false
	for _, id := range r.Carry {
		switch {
		case id != uuid.Nil && !pastMarker:
			seats = append(seats, id)
		case id == uuid.Nil && !pastMarker:
			pastMarker = true
		default:
			total++
		}
	}
	return mindsAglowAsk(ctx, seats, total+max(paid, 0))
}

// mindsAglowDraw has each player draw the total, in turn order from the
// caster.
func mindsAglowDraw(ctx *Context, total int) error {
	if total <= 0 {
		return nil
	}
	for _, id := range seatsFromController(ctx.Game, ctx.Controller()) {
		if err := (DrawCards{Player: id, N: total}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
