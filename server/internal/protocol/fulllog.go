package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// FullLogOf returns the omniscient view of g and its whole game log: the
// same fold publicLogOf runs (the same kinds, collapses and seat
// indexing), over every event the game has emitted, with no PublicLogMax
// cap. Game statistics read it once, when a game ends (ADR 0165 §2).
//
// It is not cached and costs one pass over g.Events, so it is not for a
// per-frame path. Safe to call from any goroutine: ViewOfGame and the
// fold each take the game's read lock themselves.
func FullLogOf(g *game.Game) (GameView, []LogEvent) {
	view := ViewOfGame(g)
	seatIDs := seatIDsOf(&view)
	var out []LogEvent
	g.ReadSnapshot(func() {
		f := newLogFold(seatIDs, g.EventLogGeneration())
		// A fold never makes more entries than there are events.
		f.ring = newLogRing(len(g.Events) + 1)
		f.feed(g.Events)
		out = f.entries()
	})
	return view, finishLog(out, &view)
}
