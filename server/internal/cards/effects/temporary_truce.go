package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temporary Truce — Sorcery {1}{W} (Mystery Booster 2, #5948):
//
//	"Each player may draw up to two cards. For each card less than two a
//	 player draws this way, that player gains 2 life."
//
// Each player, the active player first and then in turn order, is asked for a
// number from 0 to two (a draw limit caps it, CR 121.2b), draws that many, and
// then gains 2 life for each card short of two. The shortfall is counted from
// the draws that really arrived (a draw a replacement swallowed, a dredge,
// pays life as a card not drawn), read from the engine's draw log around the
// player's own draw. Everything after an answer is a continuation step
// (#4162), so the next player is asked once the draw has settled.
//
// No simplification.
func init() {
	temporaryTruceChosen = NumberThen("number/temporary-truce-draw-up-to-two",
		func(ctx *Context, r game.NumberAnswer) error {
			if len(r.Carry) == 0 {
				return nil
			}
			ctx.RecordStep(temporaryTruceRecord, r.Carry, r.Number)
			return temporaryTruceRest.Run(ctx)
		})
	Register(Spec{
		OracleID:     "ffe6371c-d137-4062-8c6e-6e9794ab25bc",
		Name:         "Temporary Truce",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return temporaryTruceAsk(ctx, gapB34TableInTurnOrder(ctx))
		},
	})
}

// temporaryTruceAsk asks the first of `order` how many cards to draw; the rest
// of the table rides the prompt.
func temporaryTruceAsk(ctx *Context, order []uuid.UUID) error {
	for len(order) > 0 && ctx.PlayerByID(order[0]) == nil {
		order = order[1:]
	}
	if len(order) == 0 {
		return nil
	}
	who := order[0]
	return ChooseNumber{
		Chooser:  who,
		Min:      0,
		Max:      ctx.Game.DrawChoiceCeilingForEffect(who, 2),
		Question: "Temporary Truce — draw up to two cards (gain 2 life for each card less than two)",
		Unit:     game.PayAmountCards,
		Goal:     func(*Context) int { return 2 },
		Then:     temporaryTruceChosen,
		Carry:    order,
	}.Apply(ctx)
}

const (
	temporaryTruceRecord   = "temporary-truce/chosen"
	temporaryTruceBaseline = "temporary-truce/draws-before"
)

// temporaryTruceChosen files who chose and how many, and runs the draw, the
// life and the next question as steps.
//
// It is assigned in init: the steps ask the next player, which names this
// branch, and a package-level initialiser would be a cycle.
var temporaryTruceChosen game.NumberThen

var temporaryTruceRest = RegisterThenSteps("then-steps/temporary-truce-rest",
	Instruction(func(ctx *Context) error {
		r := ctx.StepRecord(temporaryTruceRecord)
		if len(r.IDs) == 0 {
			return nil
		}
		ctx.RecordStep(temporaryTruceBaseline, nil, gapB34DrawEvents(ctx.Game, r.IDs[0]))
		return DrawCards{Player: r.IDs[0], N: r.N()}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		r := ctx.StepRecord(temporaryTruceRecord)
		if len(r.IDs) == 0 {
			return nil
		}
		drawn := gapB34DrawEvents(ctx.Game, r.IDs[0]) - ctx.StepRecord(temporaryTruceBaseline).N()
		if short := 2 - drawn; short > 0 {
			return GainLife{Player: r.IDs[0], Amount: 2 * short}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		r := ctx.StepRecord(temporaryTruceRecord)
		if len(r.IDs) < 2 {
			return nil
		}
		return temporaryTruceAsk(ctx, r.IDs[1:])
	}),
)
