package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flicker.go — shared pieces for the S22 exile-and-return family.
// Kept out of helpers.go for the reason #231 set: concurrent card
// batches collide on shared helper files.
//
// Two shapes of "blink" exist and the difference is the whole reason
// delayed triggers had to land first:
//
//   - **Immediate** — "exile it, then return it" resolves in one go
//     (Y'shtola Rhul, Thassa, Restoration Angel). The `Flicker`
//     primitive.
//   - **Delayed** — "exile it. At the beginning of the next end step,
//     return it" (Waterbender's Restoration, Cosmic Intervention,
//     Phelia). The exile happens now; the return is a separate CR
//     603.7 delayed triggered ability that goes on the stack a step
//     boundary later, where anyone can respond to it.
//
// Either way the permanent comes back as a NEW OBJECT — fresh
// InstanceID, no counters, no damage, no auras, summoning-sick
// again — and re-triggers every ETB it has. That is what makes blink
// simultaneously an ETB engine and a removal answer.

// returnExiledCardsToOwners is the delayed trigger's Effect for
// every "return those cards to the battlefield under their owner's
// control at the beginning of the next end step" clause.
//
// Declared as a plain package-level func rather than a closure built
// per cast so it captures nothing at all: a delayed trigger survives
// Clone / RestoreFrom by sharing its Effect func with the snapshot,
// on the same contract StackItem.Effect documents. Everything it
// needs — which cards, whose trigger — it reads off the item it is
// handed.
//
// A card that is no longer in exile when the trigger resolves is
// skipped (ReturnFromExile no-ops on it), which is the right answer
// for the one case that produces it: something else moved the card
// on in the meantime.
func returnExiledCardsToOwners(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		if err := (ReturnFromExile{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// flickerFirstLegalTarget is the immediate-blink trigger body shared
// by every "exile [up to one] target creature [you control], then
// return that card to the battlefield under your control" — Thassa,
// Deep-Dwelling's end step and Conjurer's Closet's. The target clause
// (whose it must be, "other" or not, "up to one" or not) is the only
// thing that differs between printings and lives on the TargetSpec,
// not here.
//
// A target that left legality in response (CR 608.2b) or an "up to
// one" left unfilled is a no-op, not an error.
func flickerFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return Flicker{Target: id, Controller: item.Controller}.Apply(ctx)
}

// exileTargetsThenScheduleReturn exiles every still-legal card target
// on the resolving item as ONE event and, from that exile's
// continuation, schedules the delayed trigger that returns the cards
// which actually made it to exile, in announce order.
//
// The IDs survive the move: exiling does not re-mint an InstanceID,
// only returning to the battlefield does (CR 400.7), so the delayed
// trigger can name the exiled cards by the ID the sweep saw.
//
// #870: the payload used to be built per exile call that returned no
// error, which is not the same set — a commander among the targets
// pauses on the CR 903.9 prompt, so its ID went into the payload
// before its owner had answered, and a blink of three creatures was
// not one event at all but three. The batch form is both halves at
// once: the legs leave together, and the continuation is handed the
// cards that reached exile.
func exileTargetsThenScheduleReturn(ctx *Context, label string) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		ids = append(ids, t.ID)
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract massEffect.apply explains.
	item := ctx.Item
	return ctx.Game.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		if len(exiled) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEnd,
			Label: label,
			Cards: exiled,
			Body:  returnExiledToOwnersBody,
		}.Apply(NewContext(g, item))
	})
}
