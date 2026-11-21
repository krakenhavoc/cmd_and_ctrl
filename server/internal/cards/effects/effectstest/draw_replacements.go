package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_replacements_test.go — #1222 from the card side: the CR 614
// window on the draw AMOUNT, watched through the ordinary
// Spec.Replacements slot.
//
// Two printed cards again: Thought Reflection's bare clause and
// Alhammarret's Archive's same clause with "except the first one you
// draw in each of your draw steps", plus a life half on a different
// event.

const (
	ThoughtReflectionOracle   = "692a6833-3014-42f6-b1ad-333bb3292c65"
	AlhammarretsArchiveOracle = "d5c4d36c-54b3-4149-906b-a57a670260fc"
)

// drawFor runs an effect-side draw of n cards for `seat` under the
// write lock and reports how many cards reached the hand.
func DrawFor(t *testing.T, g *game.Game, seat *game.Player, n int) int {
	t.Helper()
	before := seat.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(seat.ID, n); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	return seat.Hand.Size() - before
}
