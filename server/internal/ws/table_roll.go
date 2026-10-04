package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// table_roll.go is ADR 0121 §5's rate limit: the hub accepts one
// roll_table_die per seat per tableRollInterval and refuses the next
// with tableRollTooSoon. The public log keeps 200 entries
// (protocol.PublicLogMax), and a held-down button must not push the
// game's history out of it.

// tableRollInterval is how long a seat waits between two table rolls.
// The client disables its own menu items for the same 2 s.
const tableRollInterval = 2 * time.Second

// tableRollTooSoon is the refusal a seat gets inside the interval.
const tableRollTooSoon = "wait for your last roll to land"

// tableRollLimiter remembers when each seat last rolled at the table.
// The zero value is ready to use.
type tableRollLimiter struct {
	mu   sync.Mutex
	last map[uuid.UUID]time.Time
	// now is the clock; nil is time.Now. A test sets it.
	now func() time.Time
}

// reserve claims seat's next table roll. It returns false while the
// seat's last roll is younger than tableRollInterval. On true, the
// caller must call the returned release if the roll is then refused,
// so a roll that never happened does not hold the seat's slot.
func (l *tableRollLimiter) reserve(seat uuid.UUID) (release func(), ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if l.now != nil {
		now = l.now()
	}
	prev, had := l.last[seat]
	if had && now.Sub(prev) < tableRollInterval {
		return nil, false
	}
	if l.last == nil {
		l.last = make(map[uuid.UUID]time.Time)
	}
	l.last[seat] = now
	return func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if l.last[seat].Equal(now) {
			if had {
				l.last[seat] = prev
			} else {
				delete(l.last, seat)
			}
		}
	}, true
}
