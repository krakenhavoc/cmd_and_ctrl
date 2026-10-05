package mcpseat

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2279: the viewer's own name in a log line reads "you (seat N)", the
// way targets do, so the model does not have to work it out from life totals.
func TestLogLinesMarkTheViewer(t *testing.T) {
	v := &protocol.GameView{Seats: []protocol.PlayerView{
		{ID: "a", Name: "Agent", Seat: 0},
		{ID: "b", Name: "Rival", Seat: 1},
	}}
	e := protocol.LogEvent{Text: "Agent lost 2 life; Rival wins"}
	got := logLine(e, newViewerNameWrapper(v, "a"))
	if !strings.Contains(got, "you (seat 0) lost 2 life") || !strings.Contains(got, `"Rival" wins`) {
		t.Errorf("log line: %s", got)
	}
	if strings.Contains(got, "Agent") {
		t.Errorf("the viewer is still named: %s", got)
	}
	// A wrapper built with no viewer marks nobody.
	if got := logLine(e, newNameWrapper(v)); strings.Contains(got, "you") {
		t.Errorf("marked someone as you: %s", got)
	}
}

// Two seats sharing a name (#2273) cannot be told apart in prose, so the
// entry's seat carries the mark and the name stays wrapped for both.
func TestLogLinesMarkTheViewerWhenNamesAreShared(t *testing.T) {
	v := &protocol.GameView{Seats: []protocol.PlayerView{
		{ID: "a", Name: "Agent", Seat: 0},
		{ID: "b", Name: "Agent", Seat: 1},
	}}
	nw := newViewerNameWrapper(v, "b")
	mine := logLine(protocol.LogEvent{Text: "Agent lost 2 life", Seat: 1}, nw)
	theirs := logLine(protocol.LogEvent{Text: "Agent lost 2 life", Seat: 0}, nw)
	if !strings.Contains(mine, "[seat 1 is you]") || !strings.Contains(mine, `"Agent"`) {
		t.Errorf("own line: %s", mine)
	}
	if strings.Contains(theirs, "is you") {
		t.Errorf("opponent's line marked as yours: %s", theirs)
	}
}
