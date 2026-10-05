package mcpseat

import (
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2271: another seat's action can produce a snapshot that this seat's own
// in-flight answer has not reached. In that view the seat still owes the
// same answer; opening a window on it sent the answer twice.
func TestASnapshotBeforeTheAckDoesNotAnswerTwice(t *testing.T) {
	for _, answer := range []string{"ack", "error"} {
		t.Run(answer, func(t *testing.T) {
			f := newFakeServer(t)
			s := newTestSeat(t, nil)
			joinFake(t, f, s)
			v := activeView(f)
			stale := []legal.Move{f.pass()}
			f.onAction = func(protocol.ActionPayload) string {
				// The other seat's snapshot lands first and still shows
				// the pass as owed; then the one with this seat's own
				// effect (nothing owed), then the ack or the refusal.
				f.setState(v, stale, false)
				f.setState(v, nil, false)
				return answer
			}
			f.setState(v, stale, false)
			waitFor(t, "the first answer", func() bool { return f.actionCount() >= 1 })
			if answer == "ack" {
				waitFor(t, "the answer to be counted", func() bool {
					s.mu.Lock()
					defer s.mu.Unlock()
					return s.stats.absorbed["forced"] == 1
				})
			} else {
				waitFor(t, "the refusal", func() bool {
					s.mu.Lock()
					defer s.mu.Unlock()
					return s.stats.autoErrors == 1
				})
			}
			time.Sleep(200 * time.Millisecond)
			if n := f.actionCount(); n != 1 {
				t.Errorf("actions = %d, want exactly one for one owed answer", n)
			}
			s.mu.Lock()
			defer s.mu.Unlock()
			if answer == "error" && (s.stats.shown != 0 || s.stats.absorbed["forced"] != 0) {
				t.Errorf("a refusal for a superseded window: shown=%d absorbed=%d, want 0 and 0", s.stats.shown, s.stats.absorbed["forced"])
			}
		})
	}
}
