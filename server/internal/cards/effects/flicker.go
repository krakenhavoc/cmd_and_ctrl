package effects

import (
	"strings"

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
// The cards come back as ONE entry (#1872), so each sees the others
// enter (CR 603.6a): two blinked Soul Wardens each gain a life for the
// other. A card that is no longer in exile when the trigger resolves is
// skipped (CR 603.7c), which is the right answer for the one case that
// produces it: something else moved the card on in the meantime. A
// token never comes back (CR 111.8).
func returnExiledCardsToOwners(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return ReturnFromExileTogether{Targets: ids}.Apply(NewContext(g, item))
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
	return exileTargetsThenScheduleReturnBody(ctx, label, returnExiledToOwnersBody)
}

// exileTargetsThenScheduleReturnBody is exileTargetsThenScheduleReturn
// with the delayed return's body named: Semester's End's return gives
// each card its own counters (returnExiledWithAdditionalCountersBody).
func exileTargetsThenScheduleReturnBody(ctx *Context, label string, body game.BodyRef) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		ids = append(ids, t.ID)
	}
	return exileCardsThenScheduleReturnBody(ctx, ids, label, body)
}

// exileCardsThenScheduleReturn is exileTargetsThenScheduleReturn for
// cards the effect CHOSE rather than targeted (Yorion, Sky Nomad's
// "exile any number of other nonland permanents you own and control"):
// the same one exile, and the same delayed trigger over the cards that
// reached it.
func exileCardsThenScheduleReturn(ctx *Context, ids []uuid.UUID, label string) error {
	return exileCardsThenScheduleReturnBody(ctx, ids, label, returnExiledToOwnersBody)
}

// exileCardsThenScheduleReturnBody is the one exile and the one delayed
// trigger every "exile …, return them at the beginning of the next end
// step" shares, with the return's body named.
func exileCardsThenScheduleReturnBody(ctx *Context, ids []uuid.UUID, label string, body game.BodyRef) error {
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, withWord(ctx.ItemTail(flickerScheduleReturn), body.Key()+"|"+label))
}

// returnExiledCardsWithAdditionalCounters is Semester's End's return:
// the listed cards come back together under their owners' control, and
// each enters with an additional +1/+1 counter if it's a creature and
// an additional loyalty counter if it's a planeswalker, judged card by
// card as each would exist on the battlefield (CR 614.12, #4193). A
// creature planeswalker gets both.
func returnExiledCardsWithAdditionalCounters(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return ReturnFromExileTogether{
		Targets: ids,
		CountersIfType: []game.EntryCountersIfType{
			{CardType: "creature", Counter: game.CounterPlusOne, N: 1},
			{CardType: "planeswalker", Counter: game.CounterLoyalty, N: 1},
		},
	}.Apply(NewContext(g, item))
}

// exileTargetsThenReturnOnOwnersEndStep is "exile <targets>. Return
// that card to the battlefield under its owner's control at the
// beginning of that player's next end step" (The Eternal Wanderer's
// +1, #1538). It is exileTargetsThenScheduleReturn with the return
// bound to the OWNER's turn: one delayed trigger per owner among the
// cards that reached exile, each carrying ScheduleDelayedTrigger.TurnOf,
// so the delayed ability is still the resolving ability's controller's
// (CR 603.7d) and only the moment it fires is the owner's.
func exileTargetsThenReturnOnOwnersEndStep(ctx *Context, label string) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		ids = append(ids, t.ID)
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, withWord(ctx.ItemTail(flickerReturnOnOwnersEndStep), label))
}

// exileThisThenReturnUnderAnOpponentsControl is "Exile <this>, then
// return it to the battlefield under an opponent's control" (Sol'Kanar
// the Tainted, Zuko, Conflicted): the controller names the opponent,
// and the permanent comes back under that player's control as a new
// object (CR 400.7). Only the permanent the ability came from moves —
// one that died or was flickered in response has nothing to exile —
// and with no opponent left to name nothing happens.
func exileThisThenReturnUnderAnOpponentsControl(ctx *Context, item *game.StackItem, name string) error {
	if !sourceIsStillThisPermanent(ctx.Game, item) {
		return nil
	}
	self := item.SourceCardID
	return ChoosePlayerKeyed{
		Among:    Opponents,
		Question: name + " — choose an opponent to return it under",
		Then:     flickerUnderTheChosenOpponent,
		Carry:    []uuid.UUID{self},
	}.Apply(ctx)
}

// flickerUnderTheChosenOpponent is the opponent named, and the permanent
// (the carry) exiled and returned under their control, keyed (#4253).
var flickerUnderTheChosenOpponent = ChoosePlayerThen("option-pick/flicker-return-under-an-opponent",
	func(ctx *Context, opp uuid.UUID, carry []uuid.UUID) error {
		if opp == uuid.Nil || len(carry) == 0 {
			return nil
		}
		return Flicker{Target: carry[0], Controller: opp}.Apply(ctx)
	})

// flickerScheduleReturn schedules the end-step return of what reached exile (#4253 slice 9): the return's body key and the label ride the Word.
var flickerScheduleReturn = CardTail("flicker/schedule-return-at-end-step", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled := ctx.Game, ctx.Item, o.IDs
	bodyKey, label, _ := strings.Cut(o.Tail.Word, "|")
	body, ok := game.BodyRefByKey(bodyKey)
	if !ok {
		return nil
	}

	if len(exiled) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEnd,
		Label: label,
		Cards: exiled,
		Body:  body,
	}.Apply(NewContext(g, item))
})

// flickerReturnOnOwnersEndStep schedules each owner's end-step return of what reached exile (#4253 slice 9); the label is the Word.
var flickerReturnOnOwnersEndStep = CardTail("flicker/return-on-owners-end-step", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled, label := ctx.Game, ctx.Item, o.IDs, o.Tail.Word

	// Group by owner, in the order the cards reached exile, so
	// the schedule order is stable.
	var owners []uuid.UUID
	byOwner := map[uuid.UUID][]uuid.UUID{}
	for _, id := range exiled {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.Owner == uuid.Nil {
			continue
		}
		if _, seen := byOwner[c.Owner]; !seen {
			owners = append(owners, c.Owner)
		}
		byOwner[c.Owner] = append(byOwner[c.Owner], id)
	}
	for _, owner := range owners {
		if err := (ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  label,
			TurnOf: owner,
			Cards:  byOwner[owner],
			Body:   returnExiledToOwnersBody,
		}).Apply(NewContext(g, item)); err != nil {
			return err
		}
	}
	return nil
})
