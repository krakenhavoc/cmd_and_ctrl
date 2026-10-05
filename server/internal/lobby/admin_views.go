package lobby

// admin_views.go is ADR 0124 §2 and §3.1-3.3: the admin views of
// accounts and games. GET /admin/users, /admin/users/{id},
// /admin/games and /admin/games/{id}, every one behind requireAdmin
// (http.go), so an allowlisted person in player mode gets exactly the
// 403 a non-admin gets.
//
// Each handler copies the overlay from memory first (the lobby's
// tables, or the accounts playing now), releasing each lock before the
// next, then runs the read-only store (internal/adminview, which holds
// every SQL statement), then merges the two with adminview's pure
// functions. It never holds a lobby lock while it reads the database.
//
// The live connections of a table (§3.3's `connections` and the
// `connected` counts) need the hub's sockets, which ADR 0124's PR 4
// brings (Hub.LiveSockets). Until then adminOverlay leaves the sockets
// unknown and the counts are absent, never zero.

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/adminview"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// errNoAdminViews is the four routes' answer on a server with no user
// database (ADR 0124 §2), worded like PUT /me/admin-mode's.
var errNoAdminViews = httpError(http.StatusServiceUnavailable, "admin views need the user database, and this server has none")

// adminViewsStore is the store, or the 503.
func adminViewsStore(c Config) (adminview.Store, error) {
	if c.AdminViews == nil {
		return nil, errNoAdminViews
	}
	return c.AdminViews, nil
}

// adminViewFailed logs a failed query with the error and no values
// (ADR 0124 §4) and answers 500.
func adminViewFailed(c Config, what string, err error) error {
	c.logger().Error("admin view query failed", "view", what, "err", err)
	return httpError(http.StatusInternalServerError, "could not load the admin view; try again")
}

// adminOverlay copies the lobby's tables for the merge.
func adminOverlay(c Config) adminview.Overlay {
	return adminview.Overlay{Tables: adminLiveTables(c.Lobby.LiveTables())}
}

// adminLiveTables converts the lobby's copies to the plain values the
// adminview package takes, which imports nothing of the lobby.
func adminLiveTables(in []LiveTable) []adminview.LiveTable {
	out := make([]adminview.LiveTable, 0, len(in))
	for _, t := range in {
		lt := adminview.LiveTable{
			ID:        t.ID,
			Name:      t.Name,
			State:     t.State,
			CreatedAt: t.CreatedAt,
			Archived:  t.Archived,
			Practice:  t.Practice,
			Seats:     make([]adminview.LiveSeat, 0, len(t.Seats)),
		}
		for _, s := range t.Seats {
			lt.Seats = append(lt.Seats, adminview.LiveSeat{
				Seat:        s.Seat,
				PlayerID:    s.PlayerID,
				Kind:        s.Kind,
				UserID:      s.UserID,
				Name:        s.Name,
				DisplayName: s.DisplayName,
				BotTier:     s.BotTier,
				AgentClient: s.AgentClient,
				Host:        s.Host,
				// DeckName and DiscordPending are copied once
				// lobby.LiveSeat carries them (ADR 0124's Live now PR);
				// until then a memory-only seat shows neither.
			})
		}
		out = append(out, lt)
	}
	return out
}

// adminViewAccounts is GET /admin/users (§3.1): every account, up to
// adminview.MaxAccounts. played=1d|7d|30d keeps the accounts
// cmdctrl_users_played{window} counts: a play that ended in the
// window, or a seat at a running table now.
func adminViewAccounts(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	now := c.now()
	q := adminview.AccountsQuery{Limit: adminview.MaxAccounts}
	if v := r.URL.Query().Get("played"); v != "" {
		span, ok := metrics.UsersPlayedWindow(v)
		if !ok {
			return httpError(http.StatusBadRequest, "played must be 1d, 7d or 30d")
		}
		q.PlayedSince = now.Add(-span)
	}
	q.Live = c.Lobby.MetricsLiveUsers()
	res, err := store.Accounts(r.Context(), q)
	if err != nil {
		return adminViewFailed(c, "accounts", err)
	}
	return writeJSON(w, http.StatusOK, adminview.MergeAccounts(res, q.Live, now))
}

// adminViewAccount is GET /admin/users/{id} (§3.2): one account, its
// sign-in state, games, saved decks and deck requests.
func adminViewAccount(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		return httpError(http.StatusBadRequest, "invalid user id")
	}
	live := c.Lobby.MetricsLiveUsers()
	ov := adminOverlay(c)
	rows, err := store.Account(r.Context(), id)
	if errors.Is(err, adminview.ErrNotFound) {
		return httpError(http.StatusNotFound, "account not found")
	}
	if err != nil {
		return adminViewFailed(c, "account", err)
	}
	return writeJSON(w, http.StatusOK, adminview.MergeAccount(rows, live, ov, c.now()))
}

// adminViewGames is GET /admin/games (§3.3): tables newest first, filtered
// in SQL and keyset-paginated, with memory's practice tables when the
// practice filter asks for them.
func adminViewGames(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	query := r.URL.Query()
	f, err := adminview.ParseGamesFilter(query.Get)
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error())
	}
	limit, err := adminview.ParseGamesLimit(query.Get("limit"))
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error())
	}
	var after *adminview.Cursor
	if v := query.Get("cursor"); v != "" && f.Practice != adminview.PracticeOnly {
		cur, err := adminview.ParseCursor(v)
		if err != nil {
			return httpError(http.StatusBadRequest, "cursor must be a next_cursor this server served")
		}
		after = &cur
	}

	ov := adminOverlay(c)
	var res adminview.GamesResult
	if f.Practice != adminview.PracticeOnly {
		res, err = store.Games(r.Context(), adminview.GamesQuery{
			State: f.State, Archived: f.Archived, UserID: f.UserID, Limit: limit, After: after,
		})
		if err != nil {
			return adminViewFailed(c, "games", err)
		}
	}
	if f.Practice != adminview.PracticeExclude {
		if ids := ov.PracticeUserIDs(); len(ids) > 0 {
			if ov.Accounts, err = store.AccountRefs(r.Context(), ids); err != nil {
				return adminViewFailed(c, "games", err)
			}
		}
	}
	return writeJSON(w, http.StatusOK, adminview.MergeGames(res, ov, f, after == nil, c.now()))
}

// adminViewGame is GET /admin/games/{id} (§3.3): one table's row. A table
// memory holds with no row (a practice table) is served from memory; an
// ID in neither is a 404.
func adminViewGame(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	ov := adminOverlay(c)
	row, err := store.Game(r.Context(), id)
	switch {
	case err == nil:
		return writeJSON(w, http.StatusOK, adminview.GameResponse{GeneratedAt: c.now().UnixMilli(), Game: adminview.MergeGame(row, ov)})
	case !errors.Is(err, adminview.ErrNotFound):
		return adminViewFailed(c, "game", err)
	}
	for _, t := range ov.Tables {
		if t.ID != id {
			continue
		}
		var ids []string
		for _, s := range t.Seats {
			if s.UserID != "" {
				ids = append(ids, s.UserID)
			}
		}
		if len(ids) > 0 {
			if ov.Accounts, err = store.AccountRefs(r.Context(), ids); err != nil {
				return adminViewFailed(c, "game", err)
			}
		}
		return writeJSON(w, http.StatusOK, adminview.GameResponse{GeneratedAt: c.now().UnixMilli(), Game: adminview.LiveGame(t, ov)})
	}
	return httpError(http.StatusNotFound, "game not found")
}
