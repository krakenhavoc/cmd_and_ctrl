package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b455_helpers.go — shared seams of Commander 2011, batch 455 (#5574):
// Riku of Two Reflections' two "you may pay … If you do" riders.

// S82OldB455MayPayToCopyThatSpell is the effect of a cast trigger that reads
// "you may pay <cost>. If you do, copy that spell. You may choose new targets
// for the copy." (Riku of Two Reflections): the spell is held so the answer can
// arrive after it resolves, and the copy is made from last-known information
// when it has left the stack (CR 707.10, CR 608.2h). The paid branch is the one
// Reflective Golem's {2} uses.
func S82OldB455MayPayToCopyThatSpell(cost, question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		spellID := item.Trigger.Event.CardID
		g.HoldSpellForEffect(spellID)
		return MayPay{
			Chooser:  item.Controller,
			Cost:     cost,
			Question: question,
			Then:     oldB51CopyIfPaid,
			Carry:    []uuid.UUID{spellID},
		}.Apply(NewContext(g, item))
	}
}

// s82OldB455CopyEnteredIfPaid is the paid branch of "you may pay <cost>. If
// you do, create a token that's a copy of that creature": the creature rides
// the prompt as its Carry. An on-disk key: never renamed, never reused.
var s82OldB455CopyEnteredIfPaid = PayUnlessBranches("pay-unless/old-b455-copy-entered-creature",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return CreateTokenCopy{Controller: ctx.Controller(), Copy: r.Carry[0], N: 1}.Apply(ctx)
	}, nil)

// S82OldB455MayPayToCopyThatCreature is the effect of an enters trigger that
// reads "you may pay <cost>. If you do, create a token that's a copy of that
// creature" (Riku of Two Reflections): the entering creature is the trigger's
// event card.
func S82OldB455MayPayToCopyThatCreature(cost, question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		return MayPay{
			Chooser:  item.Controller,
			Cost:     cost,
			Question: question,
			Then:     s82OldB455CopyEnteredIfPaid,
			Carry:    []uuid.UUID{item.Trigger.Event.CardID},
		}.Apply(NewContext(g, item))
	}
}
