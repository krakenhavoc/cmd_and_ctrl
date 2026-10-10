package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entered_together.go — "one of the other creatures" in a "whenever one
// or more … enter" ability (#2733, ADR 0071 amendment 2026-10-10).
//
// A "one or more" ability fires once for a whole event batch
// (OncePerBatch, CR 603.2c), on the first EventETB of it, so the item
// carries one entering permanent. A card that then names the rest of
// them — Frantic Scapegoat's "you may suspect one of the other
// creatures" — reads the batch back out of the event log: every
// EventETB stamped with the triggering event's Batch. The log is the
// game's own record (Game.Events, carried by the snapshot), so a
// restored table resolves the same choice.
//
// Append-only: a new reading of the batch is a new function.

// enteredPermanent is one permanent of an entry batch: the object, and
// the player who controlled it as it entered.
type enteredPermanent struct {
	ID         uuid.UUID
	Controller uuid.UUID
}

// enteredInTriggeringBatch lists the permanents that entered the
// battlefield in the same event batch as the item's triggering
// EventETB, in the order they entered, keeping only those still on the
// battlefield as the object that entered (CR 400.7: one that left, or
// left and came back, is a new object and not "one of" them). Nil for an
// item with no triggering EventETB.
//
// Controller is who controlled the permanent as it entered: the player
// who lost it in the first control change after its entry, when there
// is one, and otherwise its controller now. That is what "creatures you
// control enter" counted, so a creature stolen in response is still one
// of yours that entered, and one you took after it entered is not.
//
// Caller must hold g.mu.
func enteredInTriggeringBatch(ctx *Context) []enteredPermanent {
	tc := ctx.Trigger()
	if tc.Event.Kind != game.EventETB || tc.Event.Batch == 0 {
		return nil
	}
	g := ctx.Game
	events := g.EventsThisTurn()
	var out []enteredPermanent
	for i, ev := range events {
		if ev.Kind != game.EventETB || ev.Batch != tc.Event.Batch || ev.CardID == uuid.Nil {
			continue
		}
		if !g.Battlefield.Contains(ev.CardID) {
			continue
		}
		var controller uuid.UUID
		again := false
		for _, later := range events[i+1:] {
			if later.CardID != ev.CardID {
				continue
			}
			if later.Kind == game.EventETB {
				again = true
				break
			}
			if later.Kind == game.EventControlChanged && controller == uuid.Nil {
				controller = later.Target
			}
		}
		if again {
			continue
		}
		if controller == uuid.Nil {
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				continue
			}
			controller = c.Controller
		}
		out = append(out, enteredPermanent{ID: ev.CardID, Controller: controller})
	}
	return out
}
