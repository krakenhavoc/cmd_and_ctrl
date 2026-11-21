package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB88YouCastFromOutsideYourHand is the Paradox condition (Doctor Who): the
// source's controller cast a spell from anywhere other than their hand, read off
// the zone the spell left.
func gapB88YouCastFromOutsideYourHand(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		ev.OldZone != "" && ev.OldZone != game.ZoneHand
}

// gapB88SuspendedWithTimeCounters is the tail after an exile that puts N time
// counters on the exiled card and gives it suspend if it has none
// (game.GainSuspendForEffect, CR 702.62a): Tail.N counters on each card that
// landed in exile and is not a land. The tail's Carry, if any, is a set of revealed
// cards that go on the bottom of the library in a random order afterwards (Sibylline
// Soothsayer's "put the rest of the revealed cards on the bottom"). An on-disk key.
var gapB88SuspendedWithTimeCounters = CardTail("gap-b88/suspended-with-time-counters", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if err := ctx.Game.GainSuspendForEffect(ctx.Controller(), id, o.Tail.N); err != nil {
			return err
		}
	}
	return gapB88BottomRevealed(ctx, o.Tail.Carry)
})

// gapB88BottomRevealed puts the revealed cards still in the library on the bottom
// in a random order.
func gapB88BottomRevealed(ctx *Context, ids []uuid.UUID) error {
	var rest []uuid.UUID
	for _, id := range ids {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, rest)
}
