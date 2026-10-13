package ws

// chat_recent.go: the last few chat lines of a table, kept so a "report
// player" (#2984) can attach what was being said.
//
// Chat is otherwise not stored at all (protocol.ChatPayload: it bypasses
// Room.Apply, no snapshot, no dump), and this does not change that for a
// table nobody reports. The ring lives on the Room, in memory only, holds
// at most RecentChatLimit lines, and is gone when the room is. It is
// copied into the report row only at the moment someone files a report,
// which is what the privacy notice says about it.

import (
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// RecentChatLimit is how many lines a room remembers.
const RecentChatLimit = 20

// chatRing is a fixed-size ring of the most recent chat lines.
type chatRing struct {
	mu    sync.Mutex
	lines []protocol.ChatPayload // oldest first, at most RecentChatLimit
}

// RecordChat remembers one human chat line, already masked by the filter
// and stamped by the server. Bot lines are not recorded: they are
// announcements, not something a player said.
func (r *Room) RecordChat(p protocol.ChatPayload) {
	r.chat.mu.Lock()
	defer r.chat.mu.Unlock()
	r.chat.lines = append(r.chat.lines, p)
	if n := len(r.chat.lines); n > RecentChatLimit {
		r.chat.lines = append([]protocol.ChatPayload(nil), r.chat.lines[n-RecentChatLimit:]...)
	}
}

// RecentChat returns up to n of the most recent lines, oldest first. A
// copy: the caller may keep it.
func (r *Room) RecentChat(n int) []protocol.ChatPayload {
	r.chat.mu.Lock()
	defer r.chat.mu.Unlock()
	lines := r.chat.lines
	if n < len(lines) {
		lines = lines[len(lines)-n:]
	}
	return append([]protocol.ChatPayload(nil), lines...)
}
