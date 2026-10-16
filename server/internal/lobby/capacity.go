package lobby

// capacity.go is the lobby's half of ADR 0148 (#2977): the weighted
// table-load cap, the connection check on the front door, and the
// readings the metrics and the admin Live now view show.
//
// # What is capped, and what never is
//
// The load of the server is the summed cost of its unended,
// unarchived tables (capacity.Limits.TableCost: a base per table, plus
// a weight per human seat and a larger one per bot seat). The cap is
// SOFT, and it gates only the introduction of new load:
//
//   - creating a table, and opening a practice table;
//   - seating a bot at an unstarted table (the expensive seat);
//   - a join or a spectate while the connection cap is full.
//
// It never touches a table that exists: a human claiming one of its
// seats, Start, every action in a running game, a reconnect and the
// restore pass for a table already admitted are all outside it. The
// players a cap would hurt most are the ones already mid-game, and
// "the server is full" is a message for someone who has not yet
// started.
//
// # Atomicity
//
// A create does real work (a room, a store row, invites) before the
// entry is visible in l.games, so a bare "check the load, then create"
// would let a burst of concurrent creates all pass the check. reserve
// holds the table's cost against the cap from the check until the
// entry is registered, under l.mu, so a burst cannot overshoot by more
// than the in-flight work the lock already serialises.

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// ErrServerFull is capacity.ErrServerFull, so the lobby's callers can
// match it without importing the capacity package.
var ErrServerFull = capacity.ErrServerFull

// SetCapacity installs the caps. conns reports the live WebSocket
// connection count (the hub's Count); nil means the connection cap
// cannot be checked here. Call it once at boot, BEFORE RestoreFromDisk,
// which sizes the restore against the load cap. The zero Limits is
// every cap off, which is what a Lobby has until this is called.
func (l *Lobby) SetCapacity(lim capacity.Limits, conns func() int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limits = lim
	l.connCount = conns
}

// Capacity returns the limits installed by SetCapacity.
func (l *Lobby) Capacity() capacity.Limits {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.limits
}

// seatCounts splits a table's seats into the two kinds the load
// weighs. An agent's guest seat is a person's seat for this purpose:
// it is played over a socket, not by an in-process runner.
func seatCounts(players []SeatInfo) (humans, bots int) {
	for _, s := range players {
		if s.IsBot {
			bots++
		} else {
			humans++
		}
	}
	return humans, bots
}

// entryCostLocked is what one table costs against the load cap now. An
// ended table and an archived one cost nothing: neither runs a bot nor
// takes a turn, and archiving is how an operator retires a table. It
// reads the entry's cached lifecycle state, like MetricsTables, so a
// reading takes no room or game lock.
func (l *Lobby) entryCostLocked(e *gameEntry) int {
	if e.meta.State == string(game.StateEnded) || e.meta.Archived() {
		return 0
	}
	humans, bots := seatCounts(e.meta.Players)
	return l.limits.TableCost(humans, bots)
}

// loadLocked is the summed cost of every table the lobby holds, plus
// what in-flight creates have reserved.
func (l *Lobby) loadLocked() int {
	n := l.reserved
	for _, e := range l.games {
		n += l.entryCostLocked(e)
	}
	return n
}

// reserveLocked holds cost against the cap, or refuses with
// ErrServerFull. The caller releases it with releaseLoad once the
// table is registered or the create has failed. Callers hold l.mu.
func (l *Lobby) reserveLocked(cost int, reason string) error {
	if !l.limits.Fits(l.loadLocked(), cost) {
		metrics.CapacityRefused(reason)
		return capacity.ErrServerFull
	}
	l.reserved += cost
	return nil
}

// reserveLoad is reserveLocked for a caller that does not hold l.mu.
// release gives the reservation back and is safe to call once. A caller
// that registers the table it reserved for calls releaseLocked in the
// same l.mu section that adds the entry, so the table is never counted
// twice (as an entry and as a reservation) for a concurrent create to
// trip over (#3957); it holds l.mu. Either is a no-op after the other.
func (l *Lobby) reserveLoad(cost int, reason string) (release, releaseLocked func(), err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.reserveLocked(cost, reason); err != nil {
		return nil, nil, err
	}
	released := false
	releaseLocked = func() {
		if !released {
			released = true
			l.reserved -= cost
		}
	}
	release = func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		releaseLocked()
	}
	return release, releaseLocked, nil
}

// botSeatFitsLocked refuses a bot seat that would push the load past
// the cap. Callers hold l.mu.
func (l *Lobby) botSeatFitsLocked() error {
	if l.limits.Fits(l.loadLocked(), l.limits.BotSeatWeight) {
		return nil
	}
	metrics.CapacityRefused(metrics.RefuseBotSeat)
	return capacity.ErrServerFull
}

// ConnectionsFull reports whether the connection cap is reached, for
// the front doors (join, spectate) to turn a newcomer away with the
// friendly message instead of a socket that will be refused a moment
// later. A seated player reconnecting is not asked: the hub exempts
// them, and they do not come through here.
func (l *Lobby) ConnectionsFull() bool {
	l.mu.Lock()
	lim, conns := l.limits, l.connCount
	l.mu.Unlock()
	if lim.MaxConns <= 0 || conns == nil {
		return false
	}
	if conns() >= lim.MaxConns {
		metrics.CapacityRefused(metrics.RefuseJoinConnections)
		return true
	}
	return false
}

// CapacityUsage is the load against the cap, for the admin Live now
// view. It takes l.mu and nothing else, so it adds no lock to a caller
// that holds none.
func (l *Lobby) CapacityUsage() capacity.Usage {
	l.mu.Lock()
	lim, conns := l.limits, l.connCount
	load := l.loadLocked()
	l.mu.Unlock()
	u := capacity.Usage{Load: load, LoadLimit: lim.MaxLoad, ConnLimit: lim.MaxConns}
	if conns != nil {
		u.Conns = conns()
	}
	return u
}

// CapacityReading is CapacityUsage in the metrics package's shape, so
// the lobby is the capacity collector's source.
func (l *Lobby) CapacityReading() metrics.CapacityReading {
	u := l.CapacityUsage()
	return metrics.CapacityReading{Load: u.Load, LoadLimit: u.LoadLimit, Conns: u.Conns, ConnLimit: u.ConnLimit}
}
