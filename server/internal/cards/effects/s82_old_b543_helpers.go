package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b543_helpers.go — shared bodies for Lorwyn's clash cards, S82 OLD
// batch 543 (#5593).

// ClashOpponent is "that player" of a clash's tail: the opponent the
// controller clashed with, or uuid.Nil when none was chosen (CR 701.30b).
func ClashOpponent(o game.TailOutcome) uuid.UUID {
	if len(o.IDs) == 0 {
		return uuid.Nil
	}
	return o.IDs[0]
}

// oldB543ClashCounter is "If you win, put a +1/+1 counter on this
// creature": the counter goes on the source only while it is still the same
// permanent (CR 400.7). An on-disk key: never renamed, never reused.
var oldB543ClashCounter = CardTail("old-b543/clash-counter-on-this", func(ctx *Context, o game.TailOutcome) error {
	if !WonTheClash(o) || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})

// OldB543WhenThisEntersClashForCounter is "When this creature enters, clash
// with an opponent. If you win, put a +1/+1 counter on this creature."
// (Oaken Brawler, Nath's Elite).
func OldB543WhenThisEntersClashForCounter(name string) game.TriggeredAbility {
	return WhenThisEnters(name+" — clash with an opponent; if you win, put a +1/+1 counter on it",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return ClashWithAnOpponent{Then: ctx.ItemTail(oldB543ClashCounter)}.Apply(ctx)
		})
}
