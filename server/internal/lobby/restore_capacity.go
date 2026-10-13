package lobby

// restore_capacity.go is the boot restore pass's half of the load cap
// (ADR 0148, #2977).

import (
	"errors"
	"log/slog"
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// restoreAdmission decides, given every readable restore point, which
// tables there is room for, and returns the rest to be DEFERRED. A
// deferred table's restore point stays on disk untouched and its games
// row is left as it is (RestoreFromDisk keeps it out of
// closeUnresidentRows), so a later boot, or this boot with a raised
// cap, brings it back. Without this a crash under load restores the
// load that caused it, and the server loops (#2978: 29 of 36 test
// tables came straight back on boot).
//
// Who comes back first:
//
//  1. A table with a human at it before one with only bots: a person is
//     waiting at it, and a bot table is the cheap thing to lose.
//  2. Within each, the most recently active first (the restore point's
//     capture time is the table's last committed move).
//  3. First-fit down that order: a table too big for what is left is
//     deferred, but a smaller one behind it may still fit.
//
// Ended tables, archived tables and tables with no games row cost
// nothing and always come back: an ended table runs nothing, an
// archived one has its runners stopped, and a table with no row is
// dropped by the pass anyway. With the cap off it admits everything
// (returns nil).
func (l *Lobby) restoreAdmission(log *slog.Logger) ws.RestoreAdmission {
	l.mu.Lock()
	lim := l.limits
	used := l.loadLocked()
	l.mu.Unlock()
	if !lim.LoadCapped() {
		return nil
	}
	return func(cands []ws.RestoreCandidate) map[uuid.UUID]bool {
		type paying struct {
			ws.RestoreCandidate
			cost int
		}
		var queue []paying
		for _, c := range cands {
			if c.State == game.StateEnded || l.rowRetired(c.GameID) {
				continue
			}
			queue = append(queue, paying{c, lim.TableCost(c.Humans, c.Bots)})
		}
		sort.SliceStable(queue, func(i, j int) bool {
			a, b := queue[i], queue[j]
			if (a.Humans > 0) != (b.Humans > 0) {
				return a.Humans > 0
			}
			if !a.TakenAt.Equal(b.TakenAt) {
				return a.TakenAt.After(b.TakenAt)
			}
			return a.GameID.String() < b.GameID.String()
		})
		deferred := map[uuid.UUID]bool{}
		admitted := 0
		for _, c := range queue {
			if lim.Fits(used, c.cost) {
				used += c.cost
				admitted++
				continue
			}
			deferred[c.GameID] = true
		}
		if len(deferred) > 0 {
			log.Warn("restore pass: more tables than the load cap allows; the rest stay on disk for a later boot",
				"cap", lim.MaxLoad, "load_after_restore", used, "restored", admitted, "deferred", len(deferred))
		}
		return deferred
	}
}

// rowRetired reports whether a restore point's table costs the load
// nothing for a reason the games row knows: it is archived, ended, or
// it has no row at all. A store error is not retirement; the table is
// counted.
func (l *Lobby) rowRetired(id uuid.UUID) bool {
	ctx, cancel := storeCtx()
	defer cancel()
	rec, _, err := l.store.LoadGame(ctx, id)
	if err != nil {
		return errors.Is(err, ErrStoreNotFound)
	}
	return rec.ArchivedAt != nil || rec.State == string(game.StateEnded)
}
