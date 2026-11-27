package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b544_helpers.go — shared bodies for Lorwyn's clash cards, S82 OLD
// batch 544 (#5593).

// oldB544ClashCounter is "If you win, put a +1/+1 counter on this
// creature": the counter goes on the source only while it is still the same
// permanent (CR 400.7). An on-disk key: never renamed, never reused.
var oldB544ClashCounter = CardTail("old-b544/clash-counter-on-this", func(ctx *Context, o game.TailOutcome) error {
	if !WonTheClash(o) || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})

// OldB544WhenThisEntersClashForCounter is "When this creature enters, clash
// with an opponent. If you win, put a +1/+1 counter on this creature."
// (Paperfin Rascal, Adder-Staff Boggart, Bog Hoodlums).
func OldB544WhenThisEntersClashForCounter(name string) game.TriggeredAbility {
	return WhenThisEnters(name+" — clash with an opponent; if you win, put a +1/+1 counter on it",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return ClashWithAnOpponent{Then: ctx.ItemTail(oldB544ClashCounter)}.Apply(ctx)
		})
}
