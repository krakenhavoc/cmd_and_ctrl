package game

import "time"

// idle_close.go: the log line of an idle close (ADR 0211 §4, §6). The
// lobby decides when a table is idle and closes it by archiving; the
// engine only records that it happened, so anyone who opens the table's
// history later can see why it stopped.

// The EventIdleClose labels.
const (
	IdleCloseLobby   = "idle_lobby"
	IdleCloseGame    = "idle_game"
	IdleCloseBots    = "idle_bots"
	IdleCloseWarning = "idle_warning"
)

// NoteIdleClose records an idle close, or the warning before one, in the
// log. label is one of the IdleClose labels, amount the hours the table
// had been idle (minutes for IdleCloseBots), and closesAt the time a
// warned game closes (ignored for a close). Any state is accepted: an
// unstarted table is closed too.
func (g *Game) NoteIdleClose(label string, amount int, closesAt time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	ev := Event{Kind: EventIdleClose, Label: label, Amount: amount}
	if label == IdleCloseWarning && !closesAt.IsZero() {
		ev.Call = closesAt.UTC().Format(time.RFC3339)
	}
	g.EmitEvent(ev)
}
