package lobby

// restore_deferred.go is the hot half of the restore pass's load cap
// (ADR 0148 section 7 and its 2026-10-13 amendment, #3103).
//
// The boot pass defers the tables the load cap has no room for: their
// restore points stay on disk and their rows are untouched, and until
// the next boot their players read "game not found". RestoreDeferred
// brings them back WITHOUT a restart, as load frees up, by running the
// same pass again: the same restore points, the same admission (humans
// first, newest first, first-fit) and the same adoption of each room
// into the lobby. It differs from the boot in exactly three ways, all
// of which follow from the server being up:
//
//   - A table that already has a room is skipped (ws.RestoreRoomsAdmitting
//     does it), because its restore point is a live table's, not a
//     deferred one.
//   - The admitted cost is reserved against the cap while the rooms are
//     rebuilt (Lobby.admission), because creates are being checked
//     concurrently.
//   - Rows are never closed afterwards. closeUnresidentRows is the
//     boot's tidy-up of tables nobody can reach; at runtime a row with
//     no room is a table that is deferred, or one being created.
//
// Only an admin starts it (POST /admin/capacity/restore-deferred). Not a
// timer, on purpose: see the amendment.

import (
	"log/slog"
	"net/http"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// RestoreDeferredResult is what a hot restore did.
type RestoreDeferredResult struct {
	// Restored is the tables brought back and registered.
	Restored int `json:"restored"`
	// StillDeferred is the tables still on disk because the cap has no
	// room for them yet.
	StillDeferred int `json:"still_deferred"`
	// Failed is the restore points that could not be rebuilt (a newer
	// schema, an unreadable file, a table with no usable rows). They
	// are not deferred: the pass logs each, exactly as the boot does.
	Failed int `json:"failed"`
	// Load and LoadLimit are the table load after the pass, against the
	// cap (0 is no cap), so the caller can see why the rest did not fit.
	Load      int `json:"load"`
	LoadLimit int `json:"load_limit"`
}

// RestoreDeferred restores the tables the boot left on disk, as many as
// fit under the cap now. See the file comment. It also sets
// cmdctrl_boot_restore_deferred_games to the tables still waiting, so
// the RestoreDeferredGames alert clears when the last one is back.
func (l *Lobby) RestoreDeferred(log *slog.Logger) RestoreDeferredResult {
	if log == nil {
		log = slog.Default()
	}
	var res RestoreDeferredResult
	if l.dumpDir() == "" || !l.store.Durable() {
		// No restore points, or no rows to pair them with: nothing was
		// deferred, so nothing can come back.
		return l.withUsage(res)
	}

	l.restoreMu.Lock()
	defer l.restoreMu.Unlock()

	held := 0
	outcomes := l.mgr.RestoreRoomsAdmitting(l.admission(log, &held))
	for _, o := range outcomes {
		switch {
		case o.Restored():
			if l.adoptRestored(log, o) {
				res.Restored++
			} else {
				res.Failed++
			}
		case o.Deferred:
			res.StillDeferred++
		case o.Skipped != "":
			// An ended table past retention: its file is gone, as at boot.
		default:
			res.Failed++
		}
	}
	metrics.SetBootRestoreDeferred(res.StillDeferred)
	if res.Restored > 0 || res.StillDeferred > 0 {
		log.Info("hot restore of deferred tables",
			"restored", res.Restored, "still_deferred", res.StillDeferred, "failed", res.Failed)
	}
	// Every restored table is registered now, so its cost is counted by
	// the lobby itself: give the reservation back, then read the usage.
	l.releaseHeld(held)
	return l.withUsage(res)
}

func (l *Lobby) withUsage(res RestoreDeferredResult) RestoreDeferredResult {
	u := l.CapacityUsage()
	res.Load, res.LoadLimit = u.Load, u.LoadLimit
	return res
}

// adminRestoreDeferred is POST /admin/capacity/restore-deferred, behind
// requireAdmin. It takes no body. The answer is RestoreDeferredResult
// and is 200 whether or not anything was waiting, so a click on an empty
// list is a count of zero rather than an error.
func adminRestoreDeferred(c Config, w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, c.Lobby.RestoreDeferred(c.logger()))
}
