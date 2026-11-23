package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trouble_in_pairs_skip_test.go — Trouble in Pairs' first sentence, "if
// an opponent would begin an extra turn, that player skips that turn
// instead" (#2529, CR 614.10, ADR 0059 amendment 2026-10-07). Each test
// drives the real cast and the real rotation seam, and asserts who is
// active after the turn ends, never the queue alone.

func SkippedExtraTurnEvents(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventExtraTurnSkipped {
			n++
		}
	}
	return n
}
